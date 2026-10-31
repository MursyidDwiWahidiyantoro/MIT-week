// Package models mirrors api/openapi.yaml. Field names and JSON tags are a
// contract with the frontend Zod schemas; do not rename without updating
// api/openapi.yaml and frontend/src/lib/api in the same commit.
package models

// InputType is what the user submitted.
type InputType string

const (
	InputURL  InputType = "URL"
	InputFile InputType = "FILE"
	InputText InputType = "TEXT"
)

// Status is the verdict band computed by internal/scoring.
type Status string

const (
	StatusSafe         Status = "SAFE"
	StatusSuspicious   Status = "SUSPICIOUS"
	StatusTampered     Status = "TAMPERED"
	StatusInsufficient Status = "INSUFFICIENT"
)

// Dossier is the single result object returned by every audit route.
type Dossier struct {
	SchemaVersion   int       `json:"schema_version"`
	ID              string    `json:"id"`
	Timestamp       string    `json:"timestamp"`
	InputType       InputType `json:"input_type"`
	SourceReference string    `json:"source_reference"`
	Degraded        bool      `json:"degraded"`
	IsDemo          bool      `json:"is_demo"`
	Verdict         Verdict   `json:"verdict"`
	Pillars         Pillars   `json:"pillars"`
	Tactics         []Tactic  `json:"tactics"`
	ShareToken      string    `json:"share_token"`
}

// Verdict holds the computed risk assessment. Score is nil when the evidence
// was insufficient to produce one.
type Verdict struct {
	Score             *int     `json:"score"`
	Status            Status   `json:"status"`
	ShortTitle        string   `json:"short_title"`
	Explanation       string   `json:"explanation"`
	ExplanationSimple string   `json:"explanation_simple"`
	Basis             []string `json:"basis"`
	NotChecked        []string `json:"not_checked"`
	FloorsApplied     []string `json:"floors_applied"`
}

// Pillars holds the four independent evidence sources.
type Pillars struct {
	ELA  ELAPillar  `json:"ela"`
	AI   AIPillar   `json:"ai"`
	Fact FactPillar `json:"fact"`
	Link LinkPillar `json:"link"`
}

// ELAPillar is server-side image forensics.
type ELAPillar struct {
	Available       bool      `json:"available"`
	Score           int       `json:"score"`
	Reliability     float64   `json:"reliability"`
	ReliabilityNote *string   `json:"reliability_note"`
	Anomalies       []Anomaly `json:"anomalies"`
}

// Anomaly is a normalized 0-1000 box in [ymin, xmin, ymax, xmax] order.
type Anomaly struct {
	BoundingBox   [4]int `json:"bounding_box"`
	ExtractedText string `json:"extracted_text"`
	Reason        string `json:"reason"`
}

// AIPillar is the LLM's indicative assessment.
type AIPillar struct {
	Available  bool     `json:"available"`
	Score      int      `json:"score"`
	Indicators []string `json:"indicators"`
	Caveat     string   `json:"caveat"`
}

// FactStatus is the outcome of checking the claim against allowlisted sources.
type FactStatus string

const (
	FactContradicts FactStatus = "CONTRADICTS"
	FactSupports    FactStatus = "SUPPORTS"
	FactUnverified  FactStatus = "UNVERIFIED"
)

// FactPillar is claim verification against trusted publishers.
type FactPillar struct {
	Available bool       `json:"available"`
	Claim     string     `json:"claim"`
	Status    FactStatus `json:"status"`
	Sources   []Source   `json:"sources"`
}

// Source is a verified publisher reference.
type Source struct {
	Title     string     `json:"title"`
	Publisher string     `json:"publisher"`
	URL       string     `json:"url"`
	Match     FactStatus `json:"match"`
}

// LinkPillar is URL heuristic and reputation analysis.
type LinkPillar struct {
	Available     bool         `json:"available"`
	Score         int          `json:"score"`
	FinalURL      string       `json:"final_url"`
	RedirectChain []string     `json:"redirect_chain"`
	Reasons       []LinkReason `json:"reasons"`
}

// LinkReason is one human-readable finding from link analysis.
type LinkReason struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

// Tactic is an informational persuasion pattern. It never affects the score.
type Tactic struct {
	Code   string `json:"code"`
	Phrase string `json:"phrase"`
	Source string `json:"source"`
}

// ErrorEnvelope is the only error shape the API ever returns.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries a stable code, an Indonesian message and a retry hint.
type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// Error codes shared with the frontend.
const (
	CodeInvalidURL       = "INVALID_URL"
	CodeBlockedTarget    = "BLOCKED_TARGET"
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA"
	CodePayloadTooLarge  = "PAYLOAD_TOO_LARGE"
	CodeRateLimited      = "RATE_LIMITED"
	CodeUpstreamTimeout  = "UPSTREAM_TIMEOUT"
	CodeInternal         = "INTERNAL"
)
