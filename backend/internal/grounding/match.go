package grounding

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// The classifier matched loosely on 9 Nov 2026: with Tavily it let ordinary
// business news, a site's listing pages and a fact-check about an unrelated
// claim "contradict" real news, which then read "Kabar ini dibantah sumber
// terpercaya". These two checks bound what a classifier answer may mean.

var (
	factCheckTitle = regexp.MustCompile(`(?i)^\s*(\[[^\]]{2,40}\]|hoaks|hoax|cek ?fakta|disinformasi|misinformasi|klarifikasi)`)
	factCheckPath  = regexp.MustCompile(`(?i)(cekfakta|cek-fakta|hoaks|hoax|disinformasi|misinformasi|klarifikasi)`)
	listingPath    = regexp.MustCompile(`(?i)/(institusi|tag|tags|topik|category|kategori|search|cari|author|penulis)(/|$)`)
)

// IsFactCheck reports whether a search hit is a fact-check article: a
// fact-checker's own site, a verdict tag or fact-check wording at the start of
// the title, or a fact-check section in the path. Listing pages never are.
// Only a fact-check article may contradict a claim.
func IsFactCheck(title, rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || listingPath.MatchString(u.Path) {
		return false
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if host == "turnbackhoax.id" || strings.HasSuffix(host, ".turnbackhoax.id") {
		return true
	}
	return factCheckTitle.MatchString(title) || factCheckPath.MatchString(u.Path)
}

// stopwords are function words plus the vocabulary every fact-check shares,
// so that only the claim's own subject counts as overlap.
var stopwords = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`yang dan atau di ke dari untuk dengan ini itu pada akan adalah sebagai juga tidak
		bukan sudah telah oleh dalam karena bahwa para saat kata hari tahun bulan baru lalu soal terkait tentang
		video foto gambar beredar viral klaim hoaks hoax cek fakta salah benar keliru klarifikasi disinformasi
		misinformasi menyesatkan narasi unggahan postingan berita resmi warganet media sosial
		the of and to in is are a an for on with from this that new about video photo`) {
		stopwords[w] = true
	}
}

// keyTerms are the claim's content words: lower-case, at least 4 letters,
// not stopwords, with a common Indonesian suffix trimmed from long words.
func keyTerms(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len([]rune(w)) < 4 || stopwords[w] {
			continue
		}
		if len(w) > 6 {
			for _, suf := range []string{"nya", "kan", "an"} {
				if strings.HasSuffix(w, suf) {
					w = strings.TrimSuffix(w, suf)
					break
				}
			}
		}
		out[w] = true
	}
	return out
}

// SameTopic reports whether text shares at least two of the claim's key
// terms (all of them when the claim has fewer than two): a source about
// another subject may neither support nor contradict the claim.
func SameTopic(claim, text string) bool {
	want := keyTerms(claim)
	if len(want) == 0 {
		return false
	}
	have := keyTerms(text)
	shared := 0
	for w := range want {
		if have[w] {
			shared++
		}
	}
	return shared >= min(2, len(want))
}
