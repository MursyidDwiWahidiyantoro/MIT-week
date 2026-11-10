package pipeline

import (
	"strings"

	"github.com/amba/amba/backend/internal/models"
)

// Plain-language notes for the verdict. They say what was looked at and what
// was found, never that something "could not be checked" (Sir, 10 Nov: every
// finished audit ends in a clear conclusion).

// simpleMax is the contract's limit for explanation_simple.
const simpleMax = 140

// VideoNote explains a video link whose footage could not be read: the audit
// still used the post's own words and a web search.
func VideoNote(res Result) string {
	if res.Video != nil && res.Video.Status != models.VideoOK {
		return "Video tidak bisa diunduh, jadi diperiksa lewat keterangan dan pencarian web."
	}
	return ""
}

// FactNote says what the fact check found, in one sentence; "" when no claim
// was checked.
func FactNote(res Result) string {
	f := res.Pillars.Fact
	if !f.Available {
		if res.InputType == models.InputFile {
			return "Gambar ini tidak memuat klaim yang bisa dicek."
		}
		return "Tidak ada klaim yang bisa dicek."
	}
	switch {
	case f.Status == models.FactContradicts:
		return "Ada artikel cek fakta yang membantah klaim ini."
	case len(f.Sources) == 0:
		return "Belum ada artikel cek fakta yang membahas klaim ini."
	case f.Status == models.FactSupports:
		return "Artikel terpercaya sejalan dengan klaim ini."
	default:
		return "Ada artikel terkait, tetapi belum ada yang memutuskan klaim ini."
	}
}

// safeSimple is the short verdict line for SAFE: what was checked and found
// nothing, plus the one note that tells how strong that is, within the
// contract's 140 characters.
func safeSimple(res Result) string {
	var found []string
	if res.Pillars.AI.Available {
		found = append(found, "tanda AI")
	}
	if res.Pillars.ELA.Available {
		found = append(found, "suntingan")
	}
	if res.Pillars.Fact.Available {
		found = append(found, "bantahan cek fakta")
	}
	if res.Pillars.Link.Available {
		found = append(found, "tanda bahaya pada tautan")
	}
	lead := "Tidak ditemukan tanda bahaya."
	switch len(found) {
	case 0:
	case 1:
		lead = "Tidak ditemukan " + found[0] + "."
	default:
		lead = "Tidak ditemukan " + strings.Join(found[:len(found)-1], ", ") + " atau " + found[len(found)-1] + "."
	}
	note := VideoNote(res)
	if note == "" {
		note = FactNote(res)
	}
	out := lead
	if note != "" && len([]rune(lead+" "+note)) <= simpleMax {
		out = lead + " " + note
	}
	return truncateRunes(out, simpleMax)
}
