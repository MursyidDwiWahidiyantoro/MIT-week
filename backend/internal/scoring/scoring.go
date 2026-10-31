// Package scoring implements AMBA's deterministic risk score.
//
// It is pure: no I/O, no clocks, no randomness. The LLM never produces a
// score; this package combines the four pillar signals and applies the
// floors, caps and status bands from TECH_SPEC §3.2–§3.3.
package scoring

import (
	"math"
	"slices"

	"github.com/amba/amba/backend/internal/models"
)

// Base pillar weights (TECH_SPEC §3.1).
const (
	WeightELA  = 0.30
	WeightAI   = 0.15
	WeightFact = 0.35
	WeightLink = 0.20

	// MinWeight is the smallest total effective weight that yields a numeric
	// score. Below it the result is INSUFFICIENT unless a floor fires.
	MinWeight = 0.30

	// SafeMinWeight is the total effective weight SAFE requires.
	SafeMinWeight = 0.50

	// weightEpsilon absorbs binary floating-point error in the weight sums.
	// 0.35 + 0.15 is 0.49999999999999994, which must still count as 0.50.
	weightEpsilon = 1e-9
)

// Floor codes as they appear in verdict.floors_applied (api/openapi.yaml).
const (
	FloorFactContradicted = "FACT_CONTRADICTED"
	FloorLinkMalicious    = "LINK_MALICIOUS"
	FloorELACorroborated  = "ELA_CORROBORATED"
	FloorELASoloCap       = "ELA_SOLO_CAP"
	FloorInjectionGuard   = "INJECTION_GUARD"
)

// Signal is one pillar's contribution. Unavailable pillars contribute neither
// weight nor score.
type Signal struct {
	Available   bool
	Score       int
	Reliability float64
}

// Box is a normalized [ymin, xmin, ymax, xmax] rectangle on a 0–1000 scale,
// matching the Dossier anomaly contract.
type Box [4]int

// Input is everything Compute needs. The four Signals come from the pillars;
// the remaining fields are the evidence the floors inspect.
type Input struct {
	ELA  Signal
	AI   Signal
	Fact Signal
	Link Signal

	// FactContradicted is true when at least one allowlisted source returned
	// match=CONTRADICTS. Allowlist verification happens in internal/grounding.
	FactContradicted bool
	// LinkMalicious is true on a Safe Browsing hit. A link score >= 90 is
	// detected from Link directly.
	LinkMalicious bool
	// ELAHotspots are the high-difference boxes reported by internal/forensics.
	ELAHotspots []Box
	// AIAnomalies are the LLM's text-anomaly boxes.
	AIAnomalies []Box
	// InjectionSuspected is true when the LLM output was flagged by the
	// injection detector.
	InjectionSuspected bool
}

// Compute returns the verdict fields this package owns. ShortTitle and
// Explanation are presentation copy filled by the caller.
func Compute(in Input) models.Verdict {
	v := models.Verdict{
		Basis:         []string{},
		NotChecked:    []string{},
		FloorsApplied: []string{},
	}

	wELA := effWeight(in.ELA, WeightELA)
	wAI := effWeight(in.AI, WeightAI)
	wFact := effWeight(in.Fact, WeightFact)
	wLink := effWeight(in.Link, WeightLink)
	totalW := wELA + wAI + wFact + wLink

	if in.ELA.Available {
		v.Basis = append(v.Basis, "ela")
	} else {
		v.NotChecked = append(v.NotChecked, "ela")
	}
	if in.AI.Available {
		v.Basis = append(v.Basis, "ai")
	} else {
		v.NotChecked = append(v.NotChecked, "ai")
	}
	if in.Fact.Available {
		v.Basis = append(v.Basis, "fact")
	} else {
		v.NotChecked = append(v.NotChecked, "fact")
	}
	if in.Link.Available {
		v.Basis = append(v.Basis, "link")
	} else {
		v.NotChecked = append(v.NotChecked, "link")
	}

	var score *int
	if totalW+weightEpsilon >= MinWeight {
		raw := (wELA*float64(in.ELA.Score) +
			wAI*float64(in.AI.Score) +
			wFact*float64(in.Fact.Score) +
			wLink*float64(in.Link.Score)) / totalW
		s := int(math.Floor(raw + 0.5))
		score = &s
	}

	// Floors run in TECH_SPEC §3.3 order. They can rescue an otherwise
	// INSUFFICIENT input, but never to SAFE.
	if in.FactContradicted {
		score = atLeast(score, 85)
		v.FloorsApplied = append(v.FloorsApplied, FloorFactContradicted)
	}
	if in.LinkMalicious || (in.Link.Available && in.Link.Score >= 90) {
		score = atLeast(score, 85)
		v.FloorsApplied = append(v.FloorsApplied, FloorLinkMalicious)
	}
	if in.ELA.Available && in.ELA.Score >= 70 && boxesOverlap(in.ELAHotspots, in.AIAnomalies) {
		score = atLeast(score, 70)
		v.FloorsApplied = append(v.FloorsApplied, FloorELACorroborated)
	}
	// The solo cap is recorded only when it actually lowers the score: it is a
	// numerical ceiling, not evidence, so listing it for a score already below
	// 60 would make the honest-limits copy claim a cap that never mattered.
	if len(v.FloorsApplied) == 0 && score != nil &&
		!otherPillarAtLeast50(in) && *score > 60 {
		s := 60
		score = &s
		v.FloorsApplied = append(v.FloorsApplied, FloorELASoloCap)
	}
	if in.InjectionSuspected {
		score = atLeast(score, 40)
		v.FloorsApplied = append(v.FloorsApplied, FloorInjectionGuard)
		v.Basis = append(v.Basis, "llm_untrusted")
	}

	if score != nil {
		s := *score
		if s < 0 {
			s = 0
		}
		if s > 100 {
			s = 100
		}
		score = &s
	}

	switch {
	case score == nil:
		v.Status = models.StatusInsufficient
	case *score <= 20:
		v.Status = models.StatusSafe
		if totalW+weightEpsilon < SafeMinWeight || !safePillarSet(in) {
			v.Status = models.StatusInsufficient
			score = nil
		}
	case *score <= 60:
		v.Status = models.StatusSuspicious
	default:
		v.Status = models.StatusTampered
	}

	v.Score = score
	v.ExplanationSimple = ExplanationSimple(v.Status, v.FloorsApplied)
	return v
}

// ExplanationSimple returns the deterministic lead sentence from TECH_SPEC
// §8.2. Floor-specific sentences win over the generic status sentence.
func ExplanationSimple(status models.Status, floors []string) string {
	switch {
	case slices.Contains(floors, FloorFactContradicted):
		return "Kabar ini dibantah sumber terpercaya. Jangan disebarkan."
	case slices.Contains(floors, FloorLinkMalicious):
		return "Tautan ini mirip penipuan. Jangan diklik."
	case slices.Contains(floors, FloorELACorroborated):
		return "Gambar ini terindikasi diedit. Jangan langsung percaya."
	}

	switch status {
	case models.StatusSafe:
		return "Tidak ditemukan tanda bahaya, tapi tetap bijak sebelum menyebarkan."
	case models.StatusSuspicious:
		return "Belum jelas benar atau tidak. Cek dulu sebelum menyebarkan."
	case models.StatusTampered:
		// §8.2 only defines TAMPERED sentences for floors 1–3. A weighted
		// score can reach TAMPERED without one (e.g. ELA 69 + AI 100 +
		// UNVERIFIED + link 89), so keep a calm, non-certain fallback.
		return "Beberapa tanda bahaya ditemukan. Cek dulu sebelum menyebarkan."
	default:
		return "Informasinya belum cukup untuk dipastikan."
	}
}

// overlaps reports whether an anomaly box overlaps an ELA hotspot: IoU >= 0.1
// or either box's centre lies inside the other.
func overlaps(a, b Box) bool {
	ay0, ax0, ay1, ax1 := ordered(a)
	by0, bx0, by1, bx1 := ordered(b)
	// A degenerate box (zero width or height) is not evidence of anything.
	if ay1 <= ay0 || ax1 <= ax0 || by1 <= by0 || bx1 <= bx0 {
		return false
	}

	iw := min(ax1, bx1) - max(ax0, bx0)
	ih := min(ay1, by1) - max(ay0, by0)
	if iw > 0 && ih > 0 {
		inter := float64(iw * ih)
		union := rectArea(ay0, ax0, ay1, ax1) + rectArea(by0, bx0, by1, bx1) - inter
		if union > 0 && inter/union >= 0.1 {
			return true
		}
	}

	by, bx := centre(b)
	ay, ax := centre(a)
	return containsPoint(ay0, ax0, ay1, ax1, by, bx) ||
		containsPoint(by0, bx0, by1, bx1, ay, ax)
}

// effWeight returns base*reliability for an available pillar. The contract
// says reliability is 0–1; clamping keeps a bad caller from inverting or
// inflating the weights.
func effWeight(s Signal, base float64) float64 {
	if !s.Available {
		return 0
	}
	r := s.Reliability
	if !(r >= 0) { // also rejects NaN
		r = 0
	}
	if r > 1 {
		r = 1
	}
	return base * r
}

func atLeast(score *int, floor int) *int {
	if score == nil || *score < floor {
		s := floor
		return &s
	}
	return score
}

func otherPillarAtLeast50(in Input) bool {
	return (in.AI.Available && in.AI.Score >= 50) ||
		(in.Fact.Available && in.Fact.Score >= 50) ||
		(in.Link.Available && in.Link.Score >= 50)
}

func safePillarSet(in Input) bool {
	return in.Fact.Available || (in.ELA.Available && in.Link.Available)
}

func boxesOverlap(hotspots, anomalies []Box) bool {
	for _, h := range hotspots {
		for _, a := range anomalies {
			if overlaps(h, a) {
				return true
			}
		}
	}
	return false
}

func ordered(b Box) (y0, x0, y1, x1 int) {
	y0, x0, y1, x1 = b[0], b[1], b[2], b[3]
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	return
}

func rectArea(y0, x0, y1, x1 int) float64 {
	return float64(y1-y0) * float64(x1-x0)
}

func centre(b Box) (y, x float64) {
	y0, x0, y1, x1 := ordered(b)
	return float64(y0+y1) / 2, float64(x0+x1) / 2
}

func containsPoint(y0, x0, y1, x1 int, py, px float64) bool {
	return py >= float64(y0) && py <= float64(y1) &&
		px >= float64(x0) && px <= float64(x1)
}
