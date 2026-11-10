package reply

import (
	"context"
	"strings"
	"testing"

	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/token"
)

func payload() token.Payload {
	score := 78
	return token.Payload{
		ID: "AMBA-1", TS: "2026-11-06T00:00:00Z", Status: "TAMPERED",
		Score: &score, Title: "Terindikasi Hoax",
		Srcs: []token.Source{{Pub: "Kompas", URL: "https://www.kompas.com/cekfakta/1", M: "CONTRADICTS"}},
	}
}

func TestTemplateMatrixIsComplete(t *testing.T) {
	for _, status := range []string{"TAMPERED", "SUSPICIOUS", "SAFE", "INSUFFICIENT"} {
		for _, tone := range []string{"hormat", "santai", "singkat"} {
			p := token.Payload{Status: status}
			text := Template(p, tone)
			if text == "" {
				t.Errorf("missing template for %s/%s", status, tone)
			}
			if len([]rune(text)) > MaxChars {
				t.Errorf("%s/%s template too long: %d", status, tone, len([]rune(text)))
			}
		}
	}
}

func TestUnknownStatusFallsBackToInsufficient(t *testing.T) {
	if got := Template(token.Payload{Status: "WAT"}, "hormat"); got != Templates["INSUFFICIENT"]["hormat"] {
		t.Fatalf("got %q", got)
	}
}

func TestValidateStripsUnknownURLs(t *testing.T) {
	allowed := map[string]bool{"https://www.kompas.com/cekfakta/1": true}
	got, ok := Validate("Cek di https://evil.example/phish dan https://www.kompas.com/cekfakta/1 ya.", allowed)
	if !ok {
		t.Fatal("valid draft rejected")
	}
	if strings.Contains(got, "evil.example") {
		t.Errorf("invented URL survived: %q", got)
	}
	if !strings.Contains(got, "kompas.com") {
		t.Errorf("allowed URL was stripped: %q", got)
	}
}

func TestValidateRejectsCertainty(t *testing.T) {
	for _, text := range []string{
		"Ini pasti hoax.", "Sudah terbukti bohong.", "100% palsu.", "Dijamin tidak benar.",
	} {
		if _, ok := Validate(text, nil); ok {
			t.Errorf("certainty claim accepted: %q", text)
		}
	}
}

func TestValidateRejectsShaming(t *testing.T) {
	if _, ok := Validate("Jangan bodoh, cek dulu.", nil); ok {
		t.Error("shaming draft accepted")
	}
}

func TestValidateCapsLength(t *testing.T) {
	got, ok := Validate(strings.Repeat("kata ", 100), nil)
	if !ok {
		t.Fatal("long but otherwise valid draft rejected")
	}
	if len([]rune(got)) > MaxChars {
		t.Fatalf("length = %d, want <= %d", len([]rune(got)), MaxChars)
	}
}

func TestMockClientFallsBackToTemplate(t *testing.T) {
	s := New(llm.New("http://unused", "", "", true), "http://localhost")
	d := s.Draft(context.Background(), payload(), "hormat", "v1.x.y")
	if d.Source != SourceTemplate {
		t.Errorf("source = %q, want template", d.Source)
	}
	if d.Text != Template(payload(), "hormat") {
		t.Errorf("text = %q", d.Text)
	}
}

func TestUnreachableClientFallsBackToTemplate(t *testing.T) {
	s := New(llm.New("http://127.0.0.1:1", "key", "Maru", false), "http://localhost")
	d := s.Draft(context.Background(), payload(), "santai", "v1.x.y")
	if d.Source != SourceTemplate {
		t.Errorf("source = %q, want template", d.Source)
	}
}

func TestCacheKeyNeverContainsToken(t *testing.T) {
	key := cacheKey("v1.secret.payload", "hormat")
	if strings.Contains(key, "secret") || len(key) != 64 {
		t.Fatalf("key = %q", key)
	}
}

func TestValidateToneRejectsWrongRegister(t *testing.T) {
	if _, ok := ValidateTone("Makasih infonya, aku cek dulu ya.", nil, "hormat"); ok {
		t.Error("casual draft accepted as hormat")
	}
	if _, ok := ValidateTone("Mohon izin berbagi, Bapak/Ibu. Saya cek dulu.", nil, "santai"); ok {
		t.Error("formal draft accepted as santai")
	}
	if _, ok := ValidateTone("Terima kasih. Sebaiknya dicek dulu.", nil, "hormat"); !ok {
		t.Error("neutral polite draft rejected as hormat")
	}
	if _, ok := ValidateTone("Makasih infonya. Aku cek dulu ya.", nil, "santai"); !ok {
		t.Error("casual draft rejected as santai")
	}
}

// A long draft must not end with half a URL: the user would send a broken
// link to a relative.
func TestTruncationNeverCutsAURL(t *testing.T) {
	long := strings.Repeat("Ini kalimat panjang yang menjelaskan hasil pemeriksaan. ", 6) +
		"Lihat http://localhost/hasil#t=v1.eyJpZCI6IkFNQkEtMjAyNjExMDYtWldZWSIsInRzIjoiMjAyNi0xMS0wNlQwMjo0NDoxNFoiLCJzdGF0dXMiOiJTVVNQSUNJT1VTIiwic2NvcmUiOjQ4fQ"
	got, ok := ValidateTone(long, nil, "santai")
	if !ok {
		t.Fatal("draft rejected")
	}
	if len([]rune(got)) > MaxChars {
		t.Fatalf("length = %d", len([]rune(got)))
	}
	if strings.Contains(got, "http://localhost/hasil#t=v1.eyJpZCI6IkFNQkEtMjAyNjExMDYtWldZWSIsInRzIjoiMjAyNi0xMS0wNlQwMjo0NDoxNFoiLCJzdGF0dXMiOiJTVVNQSUNJT1VTIiwic2NvcmUiOjQ4fQ") {
		return // full URL kept, fine
	}
	if strings.Contains(got, "http") {
		t.Fatalf("truncated URL left in the draft: %q", got)
	}
}

// A stripped URL must not leave "cek detailnya di" hanging at the end.
func TestStrippedURLLeavesNoDanglingConnective(t *testing.T) {
	allowed := map[string]bool{"https://www.kompas.com/cekfakta/1": true}
	got, ok := ValidateTone("Makasih infonya. Ini perlu dicek lagi, detailnya di https://evil.example/x", allowed, "santai")
	if !ok {
		t.Fatal("draft rejected")
	}
	if strings.HasSuffix(strings.ToLower(strings.TrimSpace(got)), " di") {
		t.Fatalf("dangling connective left: %q", got)
	}
	if !strings.HasSuffix(got, ".") {
		t.Fatalf("draft should end as a sentence: %q", got)
	}
}

func TestAppendShareLink(t *testing.T) {
	link := "http://localhost/hasil#t=v1.abc.def"
	got := appendShareLink("Pesan ini perlu dicek dulu", link, "santai")
	if !strings.Contains(got, link) {
		t.Fatalf("share link missing: %q", got)
	}
	if !strings.HasPrefix(got, "Pesan ini perlu dicek dulu.") {
		t.Fatalf("sentence punctuation missing: %q", got)
	}
	if out := appendShareLink("Pesan ini perlu dicek dulu.", link, "singkat"); strings.Contains(out, "http") {
		t.Fatalf("singkat tone must stay one line without a link: %q", out)
	}
	// Never exceed the contract limit.
	long := strings.Repeat("kata ", 60)
	out := appendShareLink(long, link, "hormat")
	if len([]rune(out)) > MaxChars {
		t.Fatalf("length = %d", len([]rune(out)))
	}
}
