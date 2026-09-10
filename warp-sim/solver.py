"""Deterministic ecology-2 physics in Warp; one CUDA lane owns one world.

Arrays put world index last so adjacent lanes access adjacent memory. Particle
slots are reused, while a separate ordered list preserves Go's ID scheduling.
No floating point, atomics or device RNG: both SplitMix64 streams match Go.
"""
import warp as wp


@wp.struct
class Config:
    width: int
    height: int
    entities: int
    code_limit: int
    cell_capacity: int
    energy_capacity: int
    inflow: int
    maintenance: int
    mutation: int
    matter_diffusion: int
    chemical_diffusion: int
    ecology: int


@wp.struct
class State:
    cells: wp.array3d(dtype=wp.int64)  # energy, matter, slot+1, X,Y,Z
    particles: wp.array3d(dtype=wp.int64)  # pos,energy,len,ip,flag,id,parent,target,created,generation
    code: wp.array3d(dtype=wp.int64)  # slot*code_limit+instruction, op/a/b, world
    memory: wp.array3d(dtype=wp.int64)  # 8 scratch + 8 inherited
    bonds: wp.array3d(dtype=wp.int32)  # up to four linked slots+1
    order: wp.array2d(dtype=wp.int32)
    free: wp.array2d(dtype=wp.int32)
    events: wp.array3d(dtype=wp.int64)  # op,a,b,next,sensed
    control: wp.array2d(dtype=wp.int32)  # count,free_count,active_module,next_change
    meta: wp.array2d(dtype=wp.uint64)  # tick,rng,transport_rng,next_id
    stats: wp.array2d(dtype=wp.int64)
    modules: wp.array3d(dtype=wp.int64)  # module*16+id, present/cost/batch/work/consume5/delta5/usage_index, world
    changes: wp.array3d(dtype=wp.uint64)  # change, tick/module, world
    usage: wp.array2d(dtype=wp.int64)


@wp.func
def index(n: int, size: int):
    return ((n % size) + size) % size


@wp.func
def neighbor(c: Config, pos: int, direction: int):
    x = pos % c.width
    y = pos // c.width
    d = index(direction, 4)
    if d == 0:
        y -= 1
    elif d == 1:
        x += 1
    elif d == 2:
        y += 1
    else:
        x -= 1
    return index(y, c.height) * c.width + index(x, c.width)


@wp.func
def random(s: State, w: int, stream: int, n: int):
    state = s.meta[stream, w] + wp.uint64(0x9E3779B97F4A7C15)
    s.meta[stream, w] = state
    z = state
    z = (z ^ (z >> wp.uint64(30))) * wp.uint64(0xBF58476D1CE4E5B9)
    z = (z ^ (z >> wp.uint64(27))) * wp.uint64(0x94D049BB133111EB)
    z = z ^ (z >> wp.uint64(31))
    return int(z % wp.uint64(n))


@wp.func
def dissipate(s: State, w: int, p: int, amount: wp.int64):
    s.particles[1, p, w] -= amount
    s.stats[2, w] += amount


@wp.func
def mix(s: State, w: int, field: int, a: int, b: int):
    high = a
    low = b
    if s.cells[field, high, w] < s.cells[field, low, w]:
        high = b
        low = a
    diff = s.cells[field, high, w] - s.cells[field, low, w]
    n = diff // wp.int64(2)
    if diff % wp.int64(2) != wp.int64(0):
        n += wp.int64(random(s, w, 2, 2))
    s.cells[field, high, w] -= n
    s.cells[field, low, w] += n


@wp.func
def environment(s: State, c: Config, w: int):
    tick = s.meta[0, w]
    for pos in range(c.width * c.height):
        rate = c.inflow * (c.width - pos % c.width) // c.width
        n = wp.min(wp.int64(rate), wp.int64(c.cell_capacity) - s.cells[0, pos, w])
        s.cells[0, pos, w] += n
        s.stats[1, w] += n
    matter = False
    chemical = False
    mp = int(0)
    cp = int(0)
    if c.matter_diffusion > 0:
        matter = tick % wp.uint64(c.matter_diffusion) == wp.uint64(0)
        mp = int((tick // wp.uint64(c.matter_diffusion)) % wp.uint64(4))
    if c.ecology != 0 and c.chemical_diffusion > 0:
        chemical = tick % wp.uint64(c.chemical_diffusion) == wp.uint64(0)
        cp = int((tick // wp.uint64(c.chemical_diffusion)) % wp.uint64(4))
    if matter or chemical:
        for pos in range(c.width * c.height):
            coordinate = pos % c.width
            direction = int(1)
            if mp >= 2:
                coordinate = pos // c.width
                direction = 2
            if matter and coordinate % 2 == mp % 2:
                mix(s, w, 1, pos, neighbor(c, pos, direction))
            coordinate = pos % c.width
            direction = 1
            if cp >= 2:
                coordinate = pos // c.width
                direction = 2
            if chemical and coordinate % 2 == cp % 2:
                other = neighbor(c, pos, direction)
                for field in range(3, 6):
                    mix(s, w, field, pos, other)
    if c.ecology != 0:
        for pos in range(c.width * c.height):
            n = wp.min(s.cells[5, pos, w], s.cells[0, pos, w] // wp.int64(8))
            s.cells[5, pos, w] -= n
            s.cells[3, pos, w] += n
            s.cells[0, pos, w] -= wp.int64(8) * n
            s.stats[13, w] += n


@wp.func
def destination(s: State, c: Config, w: int, p: int, direction: int, matter: bool):
    start = direction
    if direction < 0:
        start = random(s, w, 1, 4)
    result = int(-1)
    for n in range(4):
        pos = neighbor(c, int(s.particles[0, p, w]), start + n)
        if s.cells[2, pos, w] == wp.int64(0) and (not matter or s.cells[1, pos, w] > wp.int64(0)):
            result = pos
            break
        if direction >= 0:
            break
    return result


@wp.func
def target(s: State, c: Config, w: int, p: int):
    result = int(-1)
    for d in range(4):
        q = int(s.cells[2, neighbor(c, int(s.particles[0, p, w]), d), w]) - 1
        if q >= 0:
            if s.particles[5, q, w] == s.particles[7, p, w]:
                result = q
                break
    return result


@wp.func
def unlink_pair(s: State, w: int, p: int, q: int):
    for d in range(4):
        if s.bonds[d, p, w] == q + 1:
            s.bonds[d, p, w] = 0
        if s.bonds[d, q, w] == p + 1:
            s.bonds[d, q, w] = 0


@wp.func
def unlink(s: State, w: int, p: int):
    for d in range(4):
        q = s.bonds[d, p, w] - 1
        if q >= 0:
            unlink_pair(s, w, p, q)


@wp.func
def random_instruction(s: State, c: Config, w: int, offset: int):
    opcodes = int(12)
    if c.ecology != 0:
        opcodes = 17
    op = random(s, w, 1, opcodes)
    a = random(s, w, 1, 65) - 1
    b = random(s, w, 1, 129) - 1
    if op == 12:
        a = random(s, w, 1, 2)
        b = 1 + random(s, w, 1, 8)
    s.code[offset, 0, w] = wp.int64(op)
    s.code[offset, 1, w] = wp.int64(a)
    s.code[offset, 2, w] = wp.int64(b)


@wp.func
def copy_code(s: State, c: Config, w: int, p: int, q: int):
    length = int(s.particles[2, p, w])
    base = q * c.code_limit
    for j in range(length):
        for k in range(3):
            s.code[base + j, k, w] = s.code[p * c.code_limit + j, k, w]
    mutate = False
    if c.mutation > 0:
        mutate = random(s, w, 1, 1000000) < c.mutation
    if mutate:
        i = random(s, w, 1, length)
        kind = random(s, w, 1, 5)
        if kind == 0:
            random_instruction(s, c, w, base + i)
        elif kind == 1:
            if length < c.code_limit:
                j = length
                while j > i:
                    for k in range(3):
                        s.code[base + j, k, w] = s.code[base + j - 1, k, w]
                    j -= 1
                random_instruction(s, c, w, base + i)
                length += 1
        elif kind == 2:
            if length > 1:
                for j in range(i, length - 1):
                    for k in range(3):
                        s.code[base + j, k, w] = s.code[base + j + 1, k, w]
                length -= 1
        elif kind == 3:
            n = wp.min(1 + random(s, w, 1, 4), wp.min(length - i, c.code_limit - length))
            for j in range(n):
                for k in range(3):
                    s.code[base + length + j, k, w] = s.code[base + i + j, k, w]
            length += n
        else:
            n = wp.min(1 + random(s, w, 1, 4), wp.min(length - i, length - 1))
            for j in range(i, length - n):
                for k in range(3):
                    s.code[base + j, k, w] = s.code[base + j + n, k, w]
            length -= n
    s.particles[2, q, w] = wp.int64(length)
    s.stats[5, w] += wp.int64(1)


@wp.func
def convert(s: State, c: Config, w: int, p: int, reaction: int, amount: int):
    pos = int(s.particles[0, p, w])
    rid = s.control[2, w] * 16 + reaction
    custom = False
    if reaction >= 0 and reaction < 16:
        custom = s.modules[rid, 0, w] != wp.int64(0)
    n = wp.int64(0)
    if custom:
        n = wp.max(wp.int64(0), wp.min(wp.int64(amount), s.modules[rid, 2, w]))
        for i in range(5):
            available = wp.int64(0)
            capacity = wp.int64(0)
            if i < 3:
                available = s.cells[i + 3, pos, w]
                capacity = s.stats[10, w]
            elif i == 3:
                available = s.cells[0, pos, w]
                capacity = wp.int64(c.cell_capacity)
            else:
                available = s.particles[1, p, w]
                capacity = wp.int64(c.energy_capacity)
            consumed = s.modules[rid, 4 + i, w]
            delta = s.modules[rid, 9 + i, w]
            if consumed > wp.int64(0):
                n = wp.min(n, available // consumed)
            if delta > wp.int64(0):
                n = wp.min(n, (capacity - available) // delta)
        n = wp.max(wp.int64(0), n)
        for i in range(3):
            s.cells[3 + i, pos, w] += n * s.modules[rid, 9 + i, w]
        s.cells[0, pos, w] += n * s.modules[rid, 12, w]
        s.particles[1, p, w] += n * s.modules[rid, 13, w]
        s.stats[22, w] += s.modules[rid, 3, w]
        s.usage[int(s.modules[rid, 14, w]), w] += n
    elif reaction >= 0 and reaction < 2 and amount > 0:
        n = wp.max(wp.int64(0), wp.min(wp.int64(amount), wp.min(s.cells[3 + reaction, pos, w], (wp.int64(c.energy_capacity) - s.particles[1, p, w]) // wp.int64(4))))
        s.cells[3 + reaction, pos, w] -= n
        s.cells[4 + reaction, pos, w] += n
        s.particles[1, p, w] += wp.int64(4) * n
    if n == wp.int64(0):
        s.stats[20, w] += wp.int64(1)
    if reaction >= 0 and reaction < 2:
        s.stats[11 + reaction, w] += n


@wp.func
def resolve(s: State, c: Config, w: int, p: int):
    op = int(s.events[0, p, w])
    a = int(s.events[1, p, w])
    b = int(s.events[2, p, w])
    cost = wp.int64(1)
    if op == 4:
        cost = wp.int64(4)
    elif op == 5:
        cost = wp.max(wp.int64(1), s.particles[2, p, w])
    elif op == 6:
        cost = wp.int64(8)
    elif op == 12 and c.ecology != 0 and a >= 0 and a < 16:
        rid = s.control[2, w] * 16 + a
        if s.modules[rid, 0, w] != wp.int64(0):
            cost = s.modules[rid, 1, w]
    if s.particles[1, p, w] <= cost:
        s.stats[21, w] += wp.int64(1)
        dissipate(s, w, p, s.particles[1, p, w])
        return
    dissipate(s, w, p, cost)
    s.particles[3, p, w] = s.events[3, p, w]
    s.stats[7, w] += wp.int64(1)
    pos = int(s.particles[0, p, w])
    if op == 1:
        s.memory[index(b, 8), p, w] = s.events[4, p, w]
    elif op == 8:
        s.memory[index(a, 8), p, w] = wp.int64(b)
    elif op == 9:
        s.memory[index(b, 8), p, w] = s.memory[index(a, 8), p, w]
    elif op == 10:
        s.particles[4, p, w] = wp.int64(s.memory[index(a, 8), p, w] >= wp.int64(b))
    elif op == 2:
        linked = False
        for d in range(4):
            linked = linked or s.bonds[d, p, w] != 0
        if not linked:
            dest = destination(s, c, w, p, a, False)
            if dest >= 0:
                s.cells[2, pos, w] = wp.int64(0)
                s.cells[2, dest, w] = wp.int64(p + 1)
                s.particles[0, p, w] = wp.int64(dest)
    elif op == 3:
        n = wp.max(wp.int64(0), wp.min(wp.int64(a), wp.min(s.cells[0, pos, w], wp.int64(c.energy_capacity) - s.particles[1, p, w])))
        s.cells[0, pos, w] -= n
        s.particles[1, p, w] += n
        s.stats[8, w] += n
        if n == wp.int64(0):
            s.stats[19, w] += wp.int64(1)
    elif op == 4:
        s.particles[7, p, w] = wp.int64(0)
        if s.control[0, w] >= c.entities:
            s.stats[18, w] += wp.int64(1)
        elif s.particles[1, p, w] <= wp.int64(12):
            s.stats[17, w] += wp.int64(1)
        else:
            dest = destination(s, c, w, p, a, True)
            if dest < 0:
                space = False
                for d in range(4):
                    if a < 0 or index(a, 4) == d:
                        space = space or s.cells[2, neighbor(c, pos, d), w] == wp.int64(0)
                if space:
                    s.stats[16, w] += wp.int64(1)
                else:
                    s.stats[15, w] += wp.int64(1)
            else:
                s.control[1, w] -= 1
                q = s.free[s.control[1, w], w]
                for i in range(10):
                    s.particles[i, q, w] = wp.int64(0)
                for i in range(16):
                    s.memory[i, q, w] = wp.int64(0)
                for i in range(4):
                    s.bonds[i, q, w] = 0
                s.particles[0, q, w] = wp.int64(dest)
                s.particles[1, q, w] = wp.int64(12)
                s.particles[5, q, w] = wp.int64(s.meta[3, w])
                s.particles[6, q, w] = s.particles[5, p, w]
                s.particles[8, q, w] = wp.int64(s.meta[0, w])
                s.particles[9, q, w] = s.particles[9, p, w] + wp.int64(1)
                s.meta[3, w] += wp.uint64(1)
                s.particles[1, p, w] -= wp.int64(12)
                s.particles[7, p, w] = s.particles[5, q, w]
                s.cells[1, dest, w] -= wp.int64(1)
                s.cells[2, dest, w] = wp.int64(q + 1)
                s.order[s.control[0, w], w] = q
                s.control[0, w] += 1
                s.stats[4, w] += wp.int64(1)
    elif op == 5 or op == 6:
        q = target(s, c, w, p)
        if q >= 0:
            if s.particles[2, q, w] == wp.int64(0):
                if op == 5:
                    copy_code(s, c, w, p, q)
                else:
                    for i in range(8):
                        s.memory[i, q, w] = s.memory[i, p, w]
                    mutate = False
                    if c.mutation > 0:
                        mutate = random(s, w, 1, 1000000) < c.mutation
                    if mutate:
                        i = random(s, w, 1, 8)
                        s.memory[i, q, w] = wp.int64(random(s, w, 1, 257) - 128)
                    for i in range(8):
                        s.memory[8 + i, q, w] = s.memory[i, q, w]
    elif op == 7 or op == 14:
        q = target(s, c, w, p)
        if q >= 0:
            if op == 7:
                n = wp.max(wp.int64(0), wp.min(wp.int64(a), wp.min(s.particles[1, p, w] - wp.int64(1), wp.int64(c.energy_capacity) - s.particles[1, q, w])))
                s.particles[1, p, w] -= n
                s.particles[1, q, w] += n
                s.stats[9, w] += n
            else:
                n = wp.max(wp.int64(0), wp.min(wp.int64(a), wp.min(s.particles[1, q, w], wp.int64(c.energy_capacity) - s.particles[1, p, w])))
                s.particles[1, q, w] -= n
                s.particles[1, p, w] += n
                s.stats[14, w] += n
    elif op == 12 and c.ecology != 0:
        convert(s, c, w, p, a, b)
    elif op == 13:
        s.particles[7, p, w] = wp.int64(0)
        start = a
        if start < 0:
            start = random(s, w, 1, 4)
        for d in range(4):
            q = int(s.cells[2, neighbor(c, pos, start + d), w]) - 1
            if q >= 0:
                s.particles[7, p, w] = s.particles[5, q, w]
                break
            if a >= 0:
                break
    elif op == 15:
        q = target(s, c, w, p)
        if q >= 0:
            exists = False
            for d in range(4):
                exists = exists or s.bonds[d, p, w] == q + 1
            if not exists:
                for d in range(4):
                    if s.bonds[d, p, w] == 0:
                        s.bonds[d, p, w] = q + 1
                        break
                for d in range(4):
                    if s.bonds[d, q, w] == 0:
                        s.bonds[d, q, w] = p + 1
                        break
    elif op == 16:
        if a < 0:
            unlink(s, w, p)
        else:
            q = target(s, c, w, p)
            if q >= 0:
                unlink_pair(s, w, p, q)


@wp.kernel
def simulate(s: State, c: Config, ticks: int, lanes: int):
    tid = wp.tid()
    if tid % lanes != 0:
        return
    w = tid // lanes
    for _ in range(ticks):
        change = s.control[3, w]
        if change < s.changes.shape[0]:
            if s.changes[change, 0, w] == s.meta[0, w]:
                s.control[2, w] = int(s.changes[change, 1, w])
                s.control[3, w] += 1
        environment(s, c, w)
        count = s.control[0, w]
        shift = int(0)
        if count > 0:
            shift = int(s.meta[0, w] % wp.uint64(count))
        for j in range(count):
            p = s.order[j, w]
            length = int(s.particles[2, p, w])
            s.events[0, p, w] = wp.int64(-1)
            if length > 0:
                ip = index(int(s.particles[3, p, w]), length)
                base = p * c.code_limit + ip
                op = int(s.code[base, 0, w])
                a = int(s.code[base, 1, w])
                b = int(s.code[base, 2, w])
                next_ip = (ip + 1) % length
                flag = s.particles[4, p, w] != wp.int64(0)
                if op == 11 and (b == 0 or (b == 1 and flag) or (b == -1 and not flag)):
                    next_ip = index(a, length)
                sensed = s.particles[1, p, w]
                pos = int(s.particles[0, p, w])
                if a == 1:
                    sensed = s.cells[0, pos, w]
                elif a == 2:
                    sensed = s.cells[1, pos, w]
                elif a >= 3 and a <= 5:
                    sensed = s.cells[a, pos, w]
                s.events[0, p, w] = wp.int64(op)
                s.events[1, p, w] = wp.int64(a)
                s.events[2, p, w] = wp.int64(b)
                s.events[3, p, w] = wp.int64(next_ip)
                s.events[4, p, w] = sensed
        for j in range(count):
            p = s.order[(j + shift) % count, w]
            if s.events[0, p, w] >= wp.int64(0):
                resolve(s, c, w, p)
        remaining = int(0)
        for j in range(s.control[0, w]):
            p = s.order[j, w]
            dissipate(s, w, p, wp.min(s.particles[1, p, w], wp.int64(c.maintenance)))
            if s.particles[1, p, w] == wp.int64(0):
                unlink(s, w, p)
                pos = int(s.particles[0, p, w])
                s.cells[2, pos, w] = wp.int64(0)
                s.cells[1, pos, w] += wp.int64(1)
                s.free[s.control[1, w], w] = p
                s.control[1, w] += 1
                s.stats[6, w] += wp.int64(1)
            else:
                s.order[remaining, w] = p
                remaining += 1
        s.control[0, w] = remaining
        s.meta[0, w] += wp.uint64(1)
