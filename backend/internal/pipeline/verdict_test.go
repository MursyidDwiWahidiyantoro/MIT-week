package pipeline

import (
	"slices"
	"testing"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/scoring"
)

// An unverified claim is shown but is not evidence: on 9 Nov it made all
// 11 real Kompas news videos and 16 of 20 untouched price graphics SUSPICIOUS.
func TestUnverifiedClaimIsNotRisk(t *testing.T) {
	unverified := Result{InputType: models.InputURL, Pillars: models.Pillars{
		Fact: models.FactPillar{Available: true, Claim: "Gedung runtuh hari ini", Status: models.FactUnverified},
		Link: models.LinkPillar{Available: true, Score: 0},
	}}
	v := Verdict(unverified, false)
	if v.Status != models.StatusInsufficient || v.Score != nil {
		t.Fatalf("unverified claim: status %s score %v, want INSUFFICIENT with no score", v.Status, v.Score)
	}
	if !slices.Contains(v.NotChecked, "fact") {
		t.Fatalf("not_checked = %v, want fact listed", v.NotChecked)
	}

	contradicted := unverified
	contradicted.Pillars.Fact = models.FactPillar{Available: true, Claim: "Gedung runtuh hari ini", Status: models.FactContradicts,
		Sources: []models.Source{{URL: "https://turnbackhoax.id/x", Match: models.FactContradicts}}}
	if v := Verdict(contradicted, false); !slices.Contains(v.FloorsApplied, scoring.FloorFactContradicted) {
		t.Fatalf("contradicted claim: %+v", v)
	}
}
