package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/token"
)

// verify handles POST /verify. The token arrives in the body, never the query,
// so it never reaches server logs.
func (h *handlers) verify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil || body.Token == "" {
		writeJSON(w, http.StatusOK, map[string]any{"valid": false})
		return
	}

	payload, err := token.Verify(h.signingKey, body.Token)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"valid": false})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"valid": true,
		"payload": map[string]any{
			"id":     payload.ID,
			"ts":     payload.TS,
			"status": payload.Status,
			"score":  payload.Score,
			"title":  payload.Title,
			"claim":  payload.Claim,
			"srcs":   payload.Srcs,
		},
	})
}

// reply handles POST /reply: token-only input, so no raw user content is ever
// accepted here.
func (h *handlers) reply(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
		Tone  string `json:"tone"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil {
		writeError(w, errInvalidURL)
		return
	}
	if body.Tone != "hormat" && body.Tone != "santai" && body.Tone != "singkat" {
		writeError(w, errInvalidURL)
		return
	}

	payload, err := token.Verify(h.signingKey, body.Token)
	if err != nil {
		writeError(w, newAPIError(http.StatusBadRequest, models.CodeInvalidURL, "Laporan tidak sah.", false))
		return
	}

	text := templateReply(payload, body.Tone)
	writeJSON(w, http.StatusOK, map[string]string{"text": text, "source": "template"})
}

// templateReply is the deterministic fallback used whenever the model is
// unavailable, mock mode is on, or validation fails. The text is keyed by
// verdict status first: a SAFE report must not apologise as if it were hoax.
// The strings mirror configs/reply_templates.json; a test asserts parity.
func templateReply(p token.Payload, tone string) string {
	status := p.Status
	if status != "TAMPERED" && status != "SUSPICIOUS" && status != "SAFE" && status != "INSUFFICIENT" {
		status = "INSUFFICIENT"
	}
	return replyTemplates[status][tone]
}

var replyTemplates = map[string]map[string]string{
	"TAMPERED": {
		"hormat":  "Terima kasih atas informasinya. Mohon maaf, sepertinya informasi ini perlu diperiksa kembali karena belum ada sumber resmi yang mendukung.",
		"santai":  "Makasih infonya. Aku cek dulu ya, soalnya belum ada sumber resminya.",
		"singkat": "Terima kasih. Sepertinya perlu dicek dulu sebelum disebarkan.",
	},
	"SUSPICIOUS": {
		"hormat":  "Terima kasih sudah berbagi. Informasi ini sebaiknya diperiksa dulu ke sumber resmi sebelum diteruskan.",
		"santai":  "Makasih share-nya. Aku cek dulu ya sebelum ikut sebarkan.",
		"singkat": "Terima kasih. Sebaiknya dicek dulu ke sumber resmi.",
	},
	"SAFE": {
		"hormat":  "Terima kasih atas informasinya. Sejauh pemeriksaan, tidak ditemukan tanda bahaya, namun tetap baik untuk memeriksa sumber resminya.",
		"santai":  "Makasih infonya. Sejauh ini nggak ada tanda bahaya, tapi tetap cek sumber resminya ya.",
		"singkat": "Terima kasih. Tidak ditemukan tanda bahaya.",
	},
	"INSUFFICIENT": {
		"hormat":  "Terima kasih atas informasinya. Mohon maaf, datanya belum cukup untuk dinilai, jadi sebaiknya diperiksa langsung ke sumber resmi.",
		"santai":  "Makasih infonya. Datanya belum cukup buat dinilai, mending cek langsung ke sumbernya.",
		"singkat": "Terima kasih. Datanya belum cukup untuk dinilai.",
	},
}
