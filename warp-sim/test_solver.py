"""Differential tests against the existing Go kernel, never a Python rewrite."""
import json
import copy
import hashlib
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

import warp as wp
from benchmark import Batch, ROOT, compare


class SolverParity(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        wp.init()
        cls.device = os.environ.get("WARP_TEST_DEVICE", "cuda:0")
        cls.tmp = tempfile.TemporaryDirectory(prefix="open-end-warp-")
        cls.addClassCleanup(cls.tmp.cleanup)
        cls.exe = Path(cls.tmp.name) / ("reference.exe" if os.name == "nt" else "reference")
        subprocess.run(["go", "build", "-o", str(cls.exe), "./cmd/warp-reference"], cwd=ROOT, check=True)

    def fixture(self, scenario, size, worlds, ticks, warmup=0):
        path = Path(self.tmp.name) / "fixture.json"
        subprocess.run([str(self.exe), "-scenario", scenario, "-size", str(size), "-worlds", str(worlds),
                        "-ticks", str(ticks), "-warmup", str(warmup), "-output", str(path)], cwd=ROOT, check=True)
        return json.loads(path.read_text(encoding="utf-8"))

    def test_all_opcodes_and_energy_starvation(self):
        fixture = self.fixture("opcodes", 6, 34, 6)
        batch = Batch(fixture["initial"], self.device)
        batch.run(6, chunk=3)
        compare(batch.project(), fixture["final"])

    def test_baseline_and_mutating_ecology(self):
        for scenario in ("baseline", "mutation"):
            with self.subTest(scenario=scenario):
                fixture = self.fixture(scenario, 12, 8, 1000)
                batch = Batch(fixture["initial"], self.device)
                batch.run(1000)
                compare(batch.project(), fixture["final"])

    def test_odd_grid_and_unequal_diffusion_intervals(self):
        fixture = self.fixture("odd", 7, 4, 500)
        batch = Batch(fixture["initial"], self.device)
        batch.run(500)
        compare(batch.project(), fixture["final"])

    def test_dsl_reload_rollback_and_chunk_boundaries(self):
        fixture = self.fixture("dsl", 12, 4, 300)
        a, b = Batch(fixture["initial"], self.device), Batch(fixture["initial"], self.device)
        a.run(300)
        b.run(100, chunk=7)
        b.run(99, chunk=3)
        b.run(101, chunk=11)
        compare(a.project(), fixture["final"])
        compare(b.project(), fixture["final"])

    def test_new_dsl_reaction_consumes_particle_energy(self):
        fixture = self.fixture("dsl", 12, 1, 0)
        world = fixture["initial"][0]
        state = world["rule_state"]
        state["active"] = state["pending"][0]["module"]
        state["pending"] = []
        state["events"][0]["sha256"] = state["active"]["sha256"]
        particle = world["particles"]["1"]
        particle["code"][0]["a"] = 2
        genome = hashlib.sha256(json.dumps({"Code":particle["code"], "Memory":[0]*8}, separators=(",", ":")).encode()).hexdigest()
        record = copy.deepcopy(next(iter(world["genomes"].values())))
        record["hash"], record["code"] = genome, copy.deepcopy(particle["code"])
        world["genomes"] = {genome:record}
        origin = copy.deepcopy(next(iter(world["origins"].values())))
        origin["hash"] = genome
        world["origins"] = {genome:origin}
        particle["genome"] = particle["origin"] = genome
        source = Path(self.tmp.name) / "reaction-input.json"
        result = Path(self.tmp.name) / "reaction-reference.json"
        source.write_text(json.dumps({"initial":[world]}), encoding="utf-8")
        # Go validates the modified module and provenance before serving as oracle.
        subprocess.run([str(self.exe), "-input", str(source), "-ticks", "1", "-output", str(result)], cwd=ROOT, check=True)
        fixture = json.loads(result.read_text(encoding="utf-8"))
        batch = Batch(fixture["initial"], self.device)
        batch.run(1)
        compare(batch.project(), fixture["final"])
        self.assertGreater(sum(fixture["final"][0]["rule_state"]["usage"].values()), 0)

    def test_isolation_pair_bindings_and_slot_reuse(self):
        fixture = self.fixture("assay", 32, 4, 500, warmup=100)
        batch = Batch(fixture["initial"], self.device)
        batch.run(500)
        compare(batch.project(), fixture["final"])


if __name__ == "__main__":
    unittest.main()
