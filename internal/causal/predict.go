package causal

import (
	"fmt"
	"math"
	"slices"
)

var Models = []string{"persistence", "constant", "micro", "macro", "micro-context"}

type Score struct {
	MSE   float64 `json:"mse"`
	MAE   float64 `json:"mae"`
	Count int     `json:"observations"`
}
type Prediction struct {
	Seed      uint64             `json:"seed"`
	Group     string             `json:"group"`
	Partition string             `json:"partition"`
	From      uint64             `json:"from_tick"`
	Truth     float64            `json:"truth"`
	Values    map[string]float64 `json:"predictions"`
}
type Fold struct {
	Seed       uint64           `json:"held_out_seed"`
	Partition  string           `json:"partition"`
	TrainSeeds []uint64         `json:"training_seeds"`
	Scores     map[string]Score `json:"scores"`
}
type Evaluation struct {
	Folds       []Fold                      `json:"folds"`
	Predictions []Prediction                `json:"predictions"`
	ByPartition map[string]map[string]Score `json:"seed_mean_scores"`
}
type regression struct{ weights []float64 }
type row struct {
	x    []float64
	y, w float64
}

func fit(rows []row, lambda float64) (regression, error) {
	if len(rows) == 0 || lambda <= 0 || math.IsNaN(lambda) || math.IsInf(lambda, 0) {
		return regression{}, fmt.Errorf("invalid regression input")
	}
	n := len(rows[0].x) + 1
	a := make([][]float64, n)
	for i := range a {
		a[i] = make([]float64, n+1)
		if i > 0 {
			a[i][i] = lambda
		}
	}
	for _, r := range rows {
		if len(r.x) != n-1 || r.w <= 0 || math.IsNaN(r.w) || math.IsInf(r.w, 0) || math.IsNaN(r.y) || math.IsInf(r.y, 0) {
			return regression{}, fmt.Errorf("invalid regression row")
		}
		x := append([]float64{1}, r.x...)
		for _, v := range x {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return regression{}, fmt.Errorf("nonfinite feature")
			}
		}
		for i, v := range x {
			for j, u := range x {
				a[i][j] += r.w * v * u
			}
			a[i][n] += r.w * v * r.y
		}
	}
	for col := 0; col < n; col++ {
		pivot := col
		for i := col + 1; i < n; i++ {
			if math.Abs(a[i][col]) > math.Abs(a[pivot][col]) {
				pivot = i
			}
		}
		if math.Abs(a[pivot][col]) < 1e-14 {
			return regression{}, fmt.Errorf("singular regression")
		}
		a[col], a[pivot] = a[pivot], a[col]
		scale := a[col][col]
		for j := col; j <= n; j++ {
			a[col][j] /= scale
		}
		for i := 0; i < n; i++ {
			if i == col {
				continue
			}
			v := a[i][col]
			for j := col; j <= n; j++ {
				a[i][j] -= v * a[col][j]
			}
		}
	}
	r := regression{weights: make([]float64, n)}
	for i := range a {
		r.weights[i] = a[i][n]
	}
	return r, nil
}
func (r regression) predict(x []float64) float64 {
	v := r.weights[0]
	for i, f := range x {
		v += r.weights[i+1] * f
	}
	return max(0, min(1, v))
}
func design(s Sample, m Member, model string) []float64 {
	switch model {
	case "constant":
		return nil
	case "micro":
		return m.X
	case "macro":
		return s.X
	default:
		return append(slices.Clone(m.X), s.X...)
	}
}
func train(samples []Sample, model string, lambda float64) (regression, []uint64, error) {
	counts := map[uint64]int{}
	for _, s := range samples {
		counts[s.Seed]++
	}
	seeds := []uint64{}
	for s := range counts {
		seeds = append(seeds, s)
	}
	slices.Sort(seeds)
	rows := []row{}
	for _, s := range samples {
		weight := 1 / float64(len(seeds)*counts[s.Seed])
		if model == "macro" || model == "constant" {
			rows = append(rows, row{design(s, Member{}, model), s.Y, weight})
		} else {
			for _, m := range s.Members {
				rows = append(rows, row{design(s, m, model), m.Alive, weight / float64(len(s.Members))})
			}
		}
	}
	r, err := fit(rows, lambda)
	return r, seeds, err
}
func predict(r regression, s Sample, model string) float64 {
	if model == "persistence" {
		return 1
	}
	if model == "macro" || model == "constant" {
		return r.predict(design(s, Member{}, model))
	}
	v := 0.0
	for _, m := range s.Members {
		v += r.predict(design(s, m, model))
	}
	return v / float64(len(s.Members))
}
func validateSamples(samples []Sample) error {
	if len(samples) == 0 {
		return fmt.Errorf("no forecast observations")
	}
	seen := map[string]bool{}
	for _, s := range samples {
		key := fmt.Sprintf("%d/%s/%s/%d", s.Seed, s.Partition, s.Group, s.From)
		if seen[key] || s.Partition == "" || s.To <= s.From || len(s.Members) < 2 || len(s.X) != len(MacroFeatures) || math.IsNaN(s.Y) || s.Y < 0 || s.Y > 1 {
			return fmt.Errorf("invalid forecast sample")
		}
		seen[key] = true
		ids := map[uint64]bool{}
		for _, x := range s.X {
			if math.IsNaN(x) || math.IsInf(x, 0) {
				return fmt.Errorf("nonfinite macro feature")
			}
		}
		sum := 0.0
		for _, m := range s.Members {
			for _, x := range m.X {
				if math.IsNaN(x) || math.IsInf(x, 0) {
					return fmt.Errorf("nonfinite micro feature")
				}
			}
			if m.ID == 0 || ids[m.ID] || len(m.X) != len(MicroFeatures) || (m.Alive != 0 && m.Alive != 1) {
				return fmt.Errorf("invalid forecast member")
			}
			ids[m.ID] = true
			sum += m.Alive
		}
		if math.Abs(sum/float64(len(s.Members))-s.Y) > 1e-12 {
			return fmt.Errorf("forecast target mismatch")
		}
	}
	return nil
}

// Evaluate holds out whole seeds, keeping all repeated groups and intervals from
// the evaluated physical world out of fitting and any target-derived statistics.
func Evaluate(samples []Sample, lambda float64) (Evaluation, error) {
	out := Evaluation{Folds: []Fold{}, Predictions: []Prediction{}, ByPartition: map[string]map[string]Score{}}
	if err := validateSamples(samples); err != nil {
		return out, err
	}
	parts := []string{}
	for _, s := range samples {
		if !slices.Contains(parts, s.Partition) {
			parts = append(parts, s.Partition)
		}
	}
	slices.Sort(parts)
	for _, part := range parts {
		seeds := []uint64{}
		for _, s := range samples {
			if s.Partition == part && !slices.Contains(seeds, s.Seed) {
				seeds = append(seeds, s.Seed)
			}
		}
		slices.Sort(seeds)
		if len(seeds) < 3 {
			return out, fmt.Errorf("need at least three seeds with samples per partition")
		}
		out.ByPartition[part] = map[string]Score{}
		for _, seed := range seeds {
			training, test := []Sample{}, []Sample{}
			for _, s := range samples {
				if s.Partition != part {
					continue
				}
				if s.Seed == seed {
					test = append(test, s)
				} else {
					training = append(training, s)
				}
			}
			fold := Fold{Seed: seed, Partition: part, Scores: map[string]Score{}}
			fitted := map[string]regression{}
			for _, model := range Models {
				if model == "persistence" {
					continue
				}
				r, trainSeeds, err := train(training, model, lambda)
				if err != nil {
					return out, err
				}
				fitted[model] = r
				fold.TrainSeeds = trainSeeds
			}
			for _, s := range test {
				p := Prediction{Seed: seed, Group: s.Group, Partition: part, From: s.From, Truth: s.Y, Values: map[string]float64{}}
				for _, model := range Models {
					v := predict(fitted[model], s, model)
					p.Values[model] = v
					e := v - s.Y
					sc := fold.Scores[model]
					sc.Count++
					sc.MSE += e * e
					sc.MAE += math.Abs(e)
					fold.Scores[model] = sc
				}
				out.Predictions = append(out.Predictions, p)
			}
			for _, model := range Models {
				sc := fold.Scores[model]
				sc.MSE /= float64(sc.Count)
				sc.MAE /= float64(sc.Count)
				fold.Scores[model] = sc
				avg := out.ByPartition[part][model]
				avg.Count += sc.Count
				avg.MSE += sc.MSE / float64(len(seeds))
				avg.MAE += sc.MAE / float64(len(seeds))
				out.ByPartition[part][model] = avg
			}
			out.Folds = append(out.Folds, fold)
		}
	}
	return out, nil
}
