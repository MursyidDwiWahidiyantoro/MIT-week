package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/pipeline"
)

// fakeAPI is a minimal Bot API: it records sendMessage texts and serves one
// downloadable PNG.
type fakeAPI struct {
	mu       sync.Mutex
	sent     []string
	methods  []string
	updates  []Update
	served   bool
	srv      *httptest.Server
	pngBytes []byte
}

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()
	f := &fakeAPI{}
	img := image.NewRGBA(image.Rect(0, 0, 80, 80))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	f.pngBytes = buf.Bytes()
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/file/botTOKEN/") {
			_, _ = w.Write(f.pngBytes)
			return
		}
		method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		var params map[string]any
		_ = json.NewDecoder(r.Body).Decode(&params)
		f.mu.Lock()
		f.methods = append(f.methods, method)
		f.mu.Unlock()
		switch method {
		case "sendMessage":
			f.mu.Lock()
			f.sent = append(f.sent, params["text"].(string))
			f.mu.Unlock()
			_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
		case "getFile":
			_, _ = w.Write([]byte(`{"ok":true,"result":{"file_path":"photos/a.png","file_size":100}}`))
		case "getUpdates":
			f.mu.Lock()
			defer f.mu.Unlock()
			if f.served {
				time.Sleep(20 * time.Millisecond)
				_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
				return
			}
			f.served = true
			raw, _ := json.Marshal(f.updates)
			_, _ = w.Write([]byte(`{"ok":true,"result":` + string(raw) + `}`))
		default:
			_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
		}
	}))
	return f
}

func (f *fakeAPI) messages() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string{}, f.sent...)
}

func score(n int) *int { return &n }

func fakeAudit(seen *[]pipeline.Request, mu *sync.Mutex) AuditFunc {
	return func(_ context.Context, req pipeline.Request) models.Dossier {
		mu.Lock()
		*seen = append(*seen, req)
		mu.Unlock()
		return models.Dossier{
			Verdict: models.Verdict{
				Status: models.StatusTampered, Score: score(85),
				ShortTitle: "Terindikasi Hoax/Manipulasi", ExplanationSimple: "Tautan ini mirip penipuan. Jangan diklik.",
			},
			Pillars: models.Pillars{Link: models.LinkPillar{
				Available: true, Score: 85, Reasons: []models.LinkReason{{Code: "LOOKALIKE_BRAND", Text: "Domain meniru nama layanan populer"}},
			}},
			Tactics:    []models.Tactic{{Code: "URGENCY", Phrase: "hari ini"}},
			ShareToken: "v1.abc.def",
		}
	}
}

func newBot(f *fakeAPI, audit AuditFunc) *Bot {
	b := New("TOKEN", "a-very-long-secret", "https://amba.example", audit)
	b.APIBase = f.srv.URL
	b.HTTP = f.srv.Client()
	return b
}

func TestWebhookRequiresSecret(t *testing.T) {
	f := newFakeAPI(t)
	defer f.srv.Close()
	var seen []pipeline.Request
	var mu sync.Mutex
	b := newBot(f, fakeAudit(&seen, &mu))

	for _, secret := range []string{"", "wrong"} {
		req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":1}`))
		if secret != "" {
			req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secret)
		}
		rec := httptest.NewRecorder()
		b.WebhookHandler().ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("secret %q: status %d", secret, rec.Code)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":1,"message":{"message_id":5,"chat":{"id":42},"text":"https://bri-verifikasi.xyz/login"}}`))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "a-very-long-secret")
	rec := httptest.NewRecorder()
	b.WebhookHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	waitFor(t, func() bool { return len(f.messages()) == 1 })
	msg := f.messages()[0]
	for _, want := range []string{"Terindikasi Hoax/Manipulasi", "skor risiko 85/100", "Domain meniru nama layanan populer", "Mendesak: \"hari ini\"", "https://amba.example/hasil#t=v1.abc.def", "bukan alat bukti hukum"} {
		if !strings.Contains(msg, want) {
			t.Errorf("reply misses %q:\n%s", want, msg)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 1 || seen[0].InputType != models.InputURL {
		t.Fatalf("audit requests = %+v", seen)
	}
}

func TestRequestKinds(t *testing.T) {
	f := newFakeAPI(t)
	defer f.srv.Close()
	var seen []pipeline.Request
	var mu sync.Mutex
	b := newBot(f, fakeAudit(&seen, &mu))
	ctx := context.Background()

	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 1}, Text: "Bantuan cair hari ini, cek https://x.example"}})
	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 2}, Text: "/cek@AmbaBot https://kompas.com/a"}})
	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 3}, Caption: "lihat ini", Photo: []PhotoSize{{FileID: "s", Width: 10, Height: 10}, {FileID: "l", Width: 80, Height: 80}}}})

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 3 {
		t.Fatalf("audits = %d", len(seen))
	}
	// A message that carries a link is checked by that link.
	if seen[0].InputType != models.InputURL || seen[0].URL != "https://x.example" || seen[1].InputType != models.InputURL || seen[1].URL != "https://kompas.com/a" {
		t.Fatalf("link requests = %+v", seen[:2])
	}
	if seen[2].InputType != models.InputFile || seen[2].ImageFormat != "png" || seen[2].OCRText != "lihat ini" {
		t.Fatalf("photo request = %+v", seen[2])
	}
}

func TestCommandsAndRefusals(t *testing.T) {
	f := newFakeAPI(t)
	defer f.srv.Close()
	var seen []pipeline.Request
	var mu sync.Mutex
	b := newBot(f, fakeAudit(&seen, &mu))
	ctx := context.Background()

	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 9}, Text: "/start"}})
	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 9}, Text: "/cek"}})
	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 9}, Text: "/hapus"}})
	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 9}, Text: "http://127.0.0.1/admin"}})
	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 9}, Document: &Document{FileID: "d", MimeType: "application/pdf"}}})
	b.Handle(ctx, Update{Message: &Message{Chat: Chat{ID: 9}, Text: "Mama minta pulsa dulu ya"}})

	got := f.messages()
	if len(got) != 6 || !strings.Contains(got[0], "Halo, ini AMBA") || !strings.Contains(got[1], "Tulis /cek") ||
		!strings.Contains(got[2], "tidak dikenal") || !strings.Contains(got[3], "tidak dapat diperiksa") || !strings.Contains(got[4], "belum didukung") ||
		!strings.Contains(got[5], "tautan dan gambar") {
		t.Fatalf("replies = %q", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 0 {
		t.Fatalf("no audit should run for commands or refused input: %+v", seen)
	}
}

func TestPerChatRateLimit(t *testing.T) {
	f := newFakeAPI(t)
	defer f.srv.Close()
	var seen []pipeline.Request
	var mu sync.Mutex
	b := newBot(f, fakeAudit(&seen, &mu))
	for i := 0; i < 5; i++ {
		b.Handle(context.Background(), Update{Message: &Message{Chat: Chat{ID: 77}, Text: "https://contoh.example/a"}})
	}
	mu.Lock()
	audits := len(seen)
	mu.Unlock()
	if audits != 3 {
		t.Fatalf("burst should allow 3 audits, ran %d", audits)
	}
	if last := f.messages()[4]; !strings.Contains(last, "Terlalu banyak") {
		t.Fatalf("last reply = %q", last)
	}
}

func TestPollProcessesUpdates(t *testing.T) {
	f := newFakeAPI(t)
	defer f.srv.Close()
	f.updates = []Update{{UpdateID: 10, Message: &Message{Chat: Chat{ID: 5}, Text: "cek https://contoh.example/b"}}}
	var seen []pipeline.Request
	var mu sync.Mutex
	b := newBot(f, fakeAudit(&seen, &mu))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Poll(ctx)
	waitFor(t, func() bool { return len(f.messages()) == 1 })
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.methods[0] != "deleteWebhook" {
		t.Fatalf("polling must clear the webhook first, methods = %v", f.methods)
	}
}

func TestFormatHandlesInsufficientAndBlocked(t *testing.T) {
	b := New("T", "s", "", nil)
	out := b.Format(models.Dossier{
		Verdict:        models.Verdict{Status: models.StatusInsufficient, ShortTitle: "Belum cukup data", ExplanationSimple: "Informasinya belum cukup untuk dipastikan."},
		SourceMetadata: &models.SourceMetadata{Status: models.MetadataBlocked},
		Degraded:       true,
	})
	for _, want := range []string{"tanpa skor", "tidak mengizinkan pembacaan otomatis", "belum lengkap"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "Laporan lengkap") {
		t.Error("no report link without a public app URL")
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}

var _ = io.EOF
