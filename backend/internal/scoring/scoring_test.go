package scoring

import (
	"slices"
	"testing"

	"github.com/amba/amba/backend/internal/models"
)

func sig(score int, reliability float64) Signal {
	return Signal{Available: true, Score: score, Reliability: reliability}
}

var (
	off  = Signal{}
	ptr0 = func(i int) *int { return &i }
)

// TestVectors recomputes every row of TECH_SPEC §3.5 from the formula.
func TestVectors(t *testing.T) {
	tests := []struct {
		name       string
		in         Input
		wantScore  *int
		wantStatus models.Status
		wantFloors []string
		wantBasis  []string
		wantMissed []string
	}{
		{
			// W=.755 raw=(26.4+9.975+35)/.755=94.536 -> 95
			name: "1 fact contradicted",
			in: Input{
				ELA: sig(88, 1.0), AI: sig(95, 0.7), Fact: sig(100, 1.0),
				FactContradicted: true,
			},
			wantScore:  ptr0(95),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorFactContradicted},
			wantBasis:  []string{"ela", "ai", "fact"},
			wantMissed: []string{"link"},
		},
		{
			// W=.425 raw=(1.5+14)/.425=36.47 -> 36
			name:       "2 unverified",
			in:         Input{AI: sig(20, 0.5), Fact: sig(40, 1.0)},
			wantScore:  ptr0(36),
			wantStatus: models.StatusSuspicious,
			wantBasis:  []string{"ai", "fact"},
			wantMissed: []string{"ela", "link"},
		},
		{
			// W=.605 raw=(12+1.05)/.605=21.57 -> 22; cap does not bind at 22
			name:       "3 supports but ela noisy",
			in:         Input{ELA: sig(80, 0.5), AI: sig(10, 0.7), Fact: sig(0, 1.0)},
			wantScore:  ptr0(22),
			wantStatus: models.StatusSuspicious,
			wantBasis:  []string{"ela", "ai", "fact"},
			wantMissed: []string{"link"},
		},
		{
			// ELA 10 is inconclusive (one-sided ELA): W=.455
			// raw=(1.05+35)/.455=79.23 -> 79 -> floor 85
			name: "4 contradicted rescues low average",
			in: Input{
				ELA: sig(10, 1.0), AI: sig(10, 0.7), Fact: sig(100, 1.0),
				FactContradicted: true,
			},
			wantScore:  ptr0(85),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorFactContradicted},
			wantBasis:  []string{"ela", "ai", "fact"},
			wantMissed: []string{"link"},
		},
		{
			// W=.30 (exactly at threshold) raw=95 -> ELA solo cap 60
			name:       "5 ela solo capped",
			in:         Input{ELA: sig(95, 1.0)},
			wantScore:  ptr0(60),
			wantStatus: models.StatusSuspicious,
			wantFloors: []string{FloorELASoloCap},
			wantBasis:  []string{"ela"},
			wantMissed: []string{"ai", "fact", "link"},
		},
		{
			name:       "6 nothing available",
			in:         Input{},
			wantStatus: models.StatusInsufficient,
			wantMissed: []string{"ela", "ai", "fact", "link"},
		},
		{
			// W=.55 raw=(0+1)/.55=1.82 -> 2, W>=.50 and Fact present -> SAFE
			name:       "7 supports and clean link",
			in:         Input{Fact: sig(0, 1.0), Link: sig(5, 1.0)},
			wantScore:  ptr0(2),
			wantStatus: models.StatusSafe,
			wantBasis:  []string{"fact", "link"},
			wantMissed: []string{"ela", "ai"},
		},
		{
			// W=.20 < .30 but S_Link>=90 raises to 85, never SAFE
			name:       "8 malicious link rescues insufficient",
			in:         Input{Link: sig(95, 1.0)},
			wantScore:  ptr0(85),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorLinkMalicious},
			wantBasis:  []string{"link"},
			wantMissed: []string{"ela", "ai", "fact"},
		},
		{
			// W=.20 < .30, S_Link 85 < 90: LINK_STRONG makes the link score the
			// score instead of INSUFFICIENT (product decision, Nov 2026).
			name:       "9 strong link alone",
			in:         Input{Link: sig(85, 1.0)},
			wantScore:  ptr0(85),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorLinkStrong},
			wantBasis:  []string{"link"},
			wantMissed: []string{"ela", "ai", "fact"},
		},
		{
			// W=.075+.20=.275 < .30: AI on a bare URL is weak, the link decides.
			name:       "10 strong link with weak ai",
			in:         Input{AI: sig(20, 0.5), Link: sig(55, 1.0)},
			wantScore:  ptr0(55),
			wantStatus: models.StatusSuspicious,
			wantFloors: []string{FloorLinkStrong},
			wantBasis:  []string{"ai", "link"},
			wantMissed: []string{"ela", "fact"},
		},
		{
			// Below LinkStrongMin the honest answer is still INSUFFICIENT.
			name:       "11 weak link alone",
			in:         Input{Link: sig(40, 1.0)},
			wantStatus: models.StatusInsufficient,
			wantBasis:  []string{"link"},
			wantMissed: []string{"ela", "ai", "fact"},
		},
		{
			// A brand-impersonating host on its own (LOOKALIKE_BRAND = 45)
			// is enough for a cautious verdict.
			name:       "11b brand lookalike alone",
			in:         Input{Link: sig(45, 1.0)},
			wantScore:  ptr0(45),
			wantStatus: models.StatusSuspicious,
			wantFloors: []string{FloorLinkStrong},
			wantBasis:  []string{"link"},
			wantMissed: []string{"ela", "ai", "fact"},
		},
		{
			// A scam pattern in the text is enough even when the model alone
			// would be too thin a basis (W = .075).
			name:       "13 scam pattern with weak ai",
			in:         Input{AI: sig(20, 0.5), ScamPattern: "CREDENTIAL_REQUEST"},
			wantScore:  ptr0(80),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorScamPattern},
			wantBasis:  []string{"ai"},
			wantMissed: []string{"ela", "fact", "link"},
		},
		{
			// The scam floor gives the score first, so LINK_STRONG (which only
			// rescues a missing score) does not stack on top.
			name:       "14 scam pattern with a lookalike link",
			in:         Input{Link: sig(45, 1.0), ScamPattern: "PRIZE_FEE"},
			wantScore:  ptr0(80),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorScamPattern},
			wantBasis:  []string{"link"},
			wantMissed: []string{"ela", "ai", "fact"},
		},
		{
			// 89 would otherwise exceed the 85 a link >= 90 gets: cap keeps the
			// score monotonic in the evidence.
			name:       "12 link-only score capped at the malicious floor",
			in:         Input{Link: sig(89, 1.0)},
			wantScore:  ptr0(85),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorLinkStrong},
			wantBasis:  []string{"link"},
			wantMissed: []string{"ela", "ai", "fact"},
		},
		{
			// One-sided ELA: W=.2+.105=.305 raw=(2+3.15)/.305=16.9 -> 17,
			// but without a fact check it is not SAFE (the old ELA+Link
			// route made this SAFE).
			name:       "15 low ela and a clean link are not safe",
			in:         Input{ELA: sig(5, 1.0), AI: sig(30, 0.7), Link: sig(10, 1.0)},
			wantStatus: models.StatusInsufficient,
			wantBasis:  []string{"ela", "ai", "link"},
			wantMissed: []string{"fact"},
		},
		{
			// W=.455 raw=(8.4+14)/.455=49.2 -> 49; the old two-sided mean
			// pulled it down to 32.
			name:       "16 low ela does not lower the score",
			in:         Input{ELA: sig(5, 1.0), AI: sig(80, 0.7), Fact: sig(40, 1.0)},
			wantScore:  ptr0(49),
			wantStatus: models.StatusSuspicious,
			wantBasis:  []string{"ela", "ai", "fact"},
			wantMissed: []string{"link"},
		},
		{
			// W=.455 raw 0; the inconclusive ELA still counts as coverage
			// (.755 >= .50) and a fact check ran, so SAFE is reachable.
			name:       "17 low ela counts as coverage for safe",
			in:         Input{ELA: sig(3, 1.0), AI: sig(0, 0.7), Fact: sig(0, 1.0)},
			wantScore:  ptr0(0),
			wantStatus: models.StatusSafe,
			wantBasis:  []string{"ela", "ai", "fact"},
			wantMissed: []string{"link"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.in)
			assertScore(t, got.Score, tt.wantScore)
			if got.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", got.Status, tt.wantStatus)
			}
			if !slices.Equal(got.FloorsApplied, tt.wantFloors) {
				t.Errorf("floors = %v, want %v", got.FloorsApplied, tt.wantFloors)
			}
			if !slices.Equal(got.Basis, tt.wantBasis) {
				t.Errorf("basis = %v, want %v", got.Basis, tt.wantBasis)
			}
			if !slices.Equal(got.NotChecked, tt.wantMissed) {
				t.Errorf("not_checked = %v, want %v", got.NotChecked, tt.wantMissed)
			}
		})
	}
}

func TestFloors(t *testing.T) {
	tests := []struct {
		name       string
		in         Input
		wantScore  *int
		wantStatus models.Status
		wantFloors []string
	}{
		{
			name:       "safe browsing hit without a high link score",
			in:         Input{Link: sig(10, 1.0), LinkMalicious: true},
			wantScore:  ptr0(85),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorLinkMalicious},
		},
		{
			name: "ela corroborated by overlapping anomaly",
			in: Input{
				ELA: sig(70, 1.0), AI: sig(0, 1.0),
				ELAHotspots: []Box{{100, 100, 200, 200}},
				AIAnomalies: []Box{{100, 100, 200, 200}},
			},
			wantScore:  ptr0(70),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorELACorroborated},
		},
		{
			name: "ela below 70 is not corroborated",
			in: Input{
				ELA: sig(69, 1.0), AI: sig(0, 1.0),
				ELAHotspots: []Box{{100, 100, 200, 200}},
				AIAnomalies: []Box{{100, 100, 200, 200}},
			},
			wantScore:  ptr0(46),
			wantStatus: models.StatusSuspicious,
		},
		{
			name: "ela high but anomaly elsewhere",
			in: Input{
				ELA: sig(70, 1.0), AI: sig(0, 1.0),
				ELAHotspots: []Box{{100, 100, 200, 200}},
				AIAnomalies: []Box{{500, 500, 600, 600}},
			},
			wantScore:  ptr0(47),
			wantStatus: models.StatusSuspicious,
		},
		{
			name:       "injection guard raises a low score",
			in:         Input{Fact: sig(0, 1.0), Link: sig(0, 1.0), InjectionSuspected: true},
			wantScore:  ptr0(40),
			wantStatus: models.StatusSuspicious,
			wantFloors: []string{FloorInjectionGuard},
		},
		{
			name:       "injection guard rescues insufficient",
			in:         Input{Link: sig(0, 1.0), InjectionSuspected: true},
			wantScore:  ptr0(40),
			wantStatus: models.StatusSuspicious,
			wantFloors: []string{FloorInjectionGuard},
		},
		{
			name: "both floors apply and the weighted score stands",
			in: Input{
				Fact: sig(100, 1.0), Link: sig(95, 1.0),
				FactContradicted: true,
			},
			wantScore:  ptr0(98),
			wantStatus: models.StatusTampered,
			wantFloors: []string{FloorFactContradicted, FloorLinkMalicious},
		},
		{
			name:       "ela solo cap only binds above 60",
			in:         Input{ELA: sig(95, 1.0), AI: sig(49, 1.0)},
			wantScore:  ptr0(60),
			wantStatus: models.StatusSuspicious,
			wantFloors: []string{FloorELASoloCap},
		},
		{
			name:       "ela solo cap skipped when another pillar reaches 50",
			in:         Input{ELA: sig(95, 1.0), AI: sig(50, 1.0)},
			wantScore:  ptr0(80),
			wantStatus: models.StatusTampered,
		},
		{
			// "no other pillar" is read relative to ELA: without ELA there is
			// no solo-ELA cap to apply.
			name:       "no cap when ela is unavailable",
			in:         Input{AI: sig(100, 1.0), Link: sig(49, 1.0)},
			wantScore:  ptr0(71),
			wantStatus: models.StatusTampered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.in)
			assertScore(t, got.Score, tt.wantScore)
			if got.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", got.Status, tt.wantStatus)
			}
			if !slices.Equal(got.FloorsApplied, tt.wantFloors) {
				t.Errorf("floors = %v, want %v", got.FloorsApplied, tt.wantFloors)
			}
		})
	}
}

func TestInjectionAddsUntrustedBasis(t *testing.T) {
	got := Compute(Input{Fact: sig(0, 1.0), Link: sig(0, 1.0), InjectionSuspected: true})
	if !slices.Contains(got.Basis, "llm_untrusted") {
		t.Fatalf("basis = %v, want llm_untrusted", got.Basis)
	}
}

func TestWeightsAndStatusThresholds(t *testing.T) {
	tests := []struct {
		name       string
		in         Input
		wantScore  *int
		wantStatus models.Status
	}{
		{
			name:       "weight just below minimum",
			in:         Input{ELA: sig(50, 0.99)},
			wantStatus: models.StatusInsufficient,
		},
		{
			name:       "weight exactly minimum",
			in:         Input{ELA: sig(50, 1.0)},
			wantScore:  ptr0(50),
			wantStatus: models.StatusSuspicious,
		},
		{
			// 0.35 + 0.15 sums to 0.49999999999999994 in float64.
			name:       "binary float weight sum still counts as exactly 0.50",
			in:         Input{Fact: sig(0, 1.0), AI: sig(0, 1.0)},
			wantScore:  ptr0(0),
			wantStatus: models.StatusSafe,
		},
		{
			name:       "safe without enough weight downgrades",
			in:         Input{AI: sig(0, 1.0), Link: sig(0, 1.0)},
			wantStatus: models.StatusInsufficient,
		},
		{
			name:       "safe without a fact check downgrades",
			in:         Input{AI: sig(0, 1.0), Link: sig(0, 1.0), ELA: off},
			wantStatus: models.StatusInsufficient,
		},
		{
			name:       "ela alone below the evidence line is insufficient",
			in:         Input{ELA: sig(49, 1.0)},
			wantStatus: models.StatusInsufficient,
		},
		{
			name:       "score 20 is safe",
			in:         Input{Fact: sig(0, 1.0), Link: sig(55, 1.0)},
			wantScore:  ptr0(20),
			wantStatus: models.StatusSafe,
		},
		{
			name:       "score 21 is suspicious",
			in:         Input{Fact: sig(0, 1.0), Link: sig(58, 1.0)},
			wantScore:  ptr0(21),
			wantStatus: models.StatusSuspicious,
		},
		{
			name:       "solo cap not recorded when it did not bind",
			in:         Input{ELA: sig(55, 1.0)},
			wantScore:  ptr0(55),
			wantStatus: models.StatusSuspicious,
		},
		{
			name:       "score 60 is suspicious",
			in:         Input{Fact: sig(60, 1.0), AI: sig(60, 1.0)},
			wantScore:  ptr0(60),
			wantStatus: models.StatusSuspicious,
		},
		{
			name:       "score 61 is tampered",
			in:         Input{Fact: sig(61, 1.0), AI: sig(61, 1.0)},
			wantScore:  ptr0(61),
			wantStatus: models.StatusTampered,
		},
		{
			name:       "reliability is clamped to one",
			in:         Input{ELA: sig(100, 2.0)},
			wantScore:  ptr0(60),
			wantStatus: models.StatusSuspicious,
		},
		{
			name:       "negative reliability contributes nothing",
			in:         Input{ELA: sig(100, -1.0)},
			wantStatus: models.StatusInsufficient,
		},
		{
			name:       "score is clamped to 100",
			in:         Input{Fact: sig(200, 1.0)},
			wantScore:  ptr0(100),
			wantStatus: models.StatusTampered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.in)
			assertScore(t, got.Score, tt.wantScore)
			if got.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", got.Status, tt.wantStatus)
			}
		})
	}
}

func TestScoreClampedToRange(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want int
	}{
		{"above 100", Input{Fact: sig(200, 1.0)}, 100},
		{"below 0", Input{AI: sig(-100, 1.0), Fact: sig(-100, 1.0)}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.in)
			if got.Score == nil || *got.Score != tt.want {
				t.Fatalf("score = %v, want %d", got.Score, tt.want)
			}
		})
	}
}

func TestRounding(t *testing.T) {
	tests := []struct {
		name      string
		in        Input
		wantScore int
	}{
		{"rounds down below half", Input{AI: sig(20, 0.5), Fact: sig(40, 1.0)}, 36},
		{"rounds up above half", Input{ELA: sig(80, 0.5), AI: sig(10, 0.7), Fact: sig(0, 1.0)}, 22},
		{"rounds up on 94.5", Input{ELA: sig(88, 1.0), AI: sig(95, 0.7), Fact: sig(100, 1.0)}, 95},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.in)
			if got.Score == nil || *got.Score != tt.wantScore {
				t.Fatalf("score = %v, want %d", got.Score, tt.wantScore)
			}
		})
	}
}

func TestExplanationSimple(t *testing.T) {
	tests := []struct {
		name   string
		status models.Status
		floors []string
		want   string
	}{
		{
			name:   "fact contradicted",
			status: models.StatusTampered,
			floors: []string{FloorFactContradicted},
			want:   "Kabar ini dibantah sumber terpercaya. Jangan disebarkan.",
		},
		{
			name:   "link malicious",
			status: models.StatusTampered,
			floors: []string{FloorLinkMalicious},
			want:   "Tautan ini mirip penipuan. Jangan diklik.",
		},
		{
			name:   "scam pattern",
			status: models.StatusTampered,
			floors: []string{FloorScamPattern},
			want:   "Pesan ini memakai pola penipuan yang umum. Jangan transfer uang atau memberi kode.",
		},
		{
			name:   "link strong",
			status: models.StatusTampered,
			floors: []string{FloorLinkStrong},
			want:   "Tautan ini punya tanda penipuan. Sebaiknya jangan diklik.",
		},
		{
			name:   "ela corroborated",
			status: models.StatusTampered,
			floors: []string{FloorELACorroborated},
			want:   "Gambar ini terindikasi diedit. Jangan langsung percaya.",
		},
		{
			name:   "other suspicious",
			status: models.StatusSuspicious,
			want:   "Belum jelas benar atau tidak. Cek dulu sebelum menyebarkan.",
		},
		{
			name:   "insufficient",
			status: models.StatusInsufficient,
			want:   "Informasinya belum cukup untuk dipastikan.",
		},
		{
			name:   "safe",
			status: models.StatusSafe,
			want:   "Tidak ditemukan tanda bahaya, tapi tetap bijak sebelum menyebarkan.",
		},
		{
			name:   "fact sentence wins over link sentence",
			status: models.StatusTampered,
			floors: []string{FloorLinkMalicious, FloorFactContradicted},
			want:   "Kabar ini dibantah sumber terpercaya. Jangan disebarkan.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExplanationSimple(tt.status, tt.floors); got != tt.want {
				t.Fatalf("ExplanationSimple(%s, %v) = %q, want %q", tt.status, tt.floors, got, tt.want)
			}
		})
	}
}

func TestComputeSetsExplanationSimple(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want string
	}{
		{
			name: "fact",
			in:   Input{Fact: sig(100, 1.0), FactContradicted: true},
			want: "Kabar ini dibantah sumber terpercaya. Jangan disebarkan.",
		},
		{
			name: "link",
			in:   Input{Link: sig(95, 1.0)},
			want: "Tautan ini mirip penipuan. Jangan diklik.",
		},
		{
			name: "ela",
			in: Input{
				ELA:         sig(80, 1.0),
				ELAHotspots: []Box{{0, 0, 100, 100}},
				AIAnomalies: []Box{{0, 0, 100, 100}},
			},
			want: "Gambar ini terindikasi diedit. Jangan langsung percaya.",
		},
		{
			name: "suspicious",
			in:   Input{AI: sig(20, 0.5), Fact: sig(40, 1.0)},
			want: "Belum jelas benar atau tidak. Cek dulu sebelum menyebarkan.",
		},
		{
			name: "insufficient",
			in:   Input{},
			want: "Informasinya belum cukup untuk dipastikan.",
		},
		{
			name: "safe",
			in:   Input{Fact: sig(0, 1.0), Link: sig(5, 1.0)},
			want: "Tidak ditemukan tanda bahaya, tapi tetap bijak sebelum menyebarkan.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compute(tt.in).ExplanationSimple; got != tt.want {
				t.Fatalf("explanation_simple = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOverlaps(t *testing.T) {
	tests := []struct {
		name string
		a, b Box
		want bool
	}{
		{"identical", Box{100, 100, 200, 200}, Box{100, 100, 200, 200}, true},
		{"iou above 0.1", Box{0, 0, 100, 100}, Box{0, 0, 50, 100}, true},
		{"small box centre inside big box", Box{0, 0, 1000, 1000}, Box{400, 400, 500, 500}, true},
		{"touching edges do not overlap", Box{0, 0, 100, 100}, Box{100, 0, 200, 100}, false},
		{"disjoint", Box{0, 0, 10, 10}, Box{20, 20, 30, 30}, false},
		{"reversed corners are normalised", Box{100, 100, 0, 0}, Box{0, 0, 100, 100}, true},
		{"zero-area hotspot is ignored", Box{100, 100, 100, 100}, Box{100, 100, 200, 200}, false},
		{"zero-area anomaly is ignored", Box{0, 0, 500, 500}, Box{100, 100, 100, 100}, false},
		{"iou exactly 0.1", Box{0, 0, 100, 100}, Box{0, 0, 20, 50}, true},
		{"iou below 0.1 and centres outside", Box{0, 0, 10, 1000}, Box{9, 0, 19, 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := overlaps(tt.a, tt.b); got != tt.want {
				t.Fatalf("Overlaps(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func assertScore(t *testing.T, got, want *int) {
	t.Helper()
	switch {
	case got == nil && want == nil:
	case got == nil:
		t.Fatalf("score = nil, want %d", *want)
	case want == nil:
		t.Fatalf("score = %d, want nil", *got)
	case *got != *want:
		t.Fatalf("score = %d, want %d", *got, *want)
	}
}
