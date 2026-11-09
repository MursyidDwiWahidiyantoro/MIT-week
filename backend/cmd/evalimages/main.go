// Command evalimages runs the real audit pipeline over folders of labelled
// images, so the team can calibrate Jejak AI and Timpa Teks with the model
// and keys they deploy with (the thresholds were set without a model key).
//
// It reads the same environment as the server (NINEROUTER_*, LLM_MOCK,
// SIGHTENGINE_*, GOOGLE_API_KEY, ...). Each argument is a folder whose
// images share one label, the folder name, for example:
//
//	evalimages ./eval/ai ./eval/asli ./eval/teks-ditempel
//
// It prints one line per image (AI score and what decided it, provenance,
// the largest pasted-text evidence, verdict and floors) and a summary per
// folder. Nothing is uploaded anywhere except to the services the server
// itself would call; nothing is written to disk.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/amba/amba/backend/internal/config"
	"github.com/amba/amba/backend/internal/httpapi"
	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/pipeline"
	"github.com/amba/amba/backend/internal/scoring"
	"github.com/amba/amba/backend/internal/security"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: evalimages <folder> [<folder> ...]")
		os.Exit(2)
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration:", err)
		os.Exit(1)
	}
	deps := httpapi.NewDeps(cfg)
	if cfg.LLMMock {
		fmt.Fprintln(os.Stderr, "note: LLM_MOCK=true, the model's answers are invented; set LLM_MOCK=false to calibrate")
	}

	fmt.Println("label\tfile\tai\tai_rel\tprovenance\tdetector\ttext_evidence\tstatus\tscore\tfloors")
	for _, dir := range os.Args[1:] {
		label := filepath.Base(filepath.Clean(dir))
		files, _ := filepath.Glob(filepath.Join(dir, "*"))
		sort.Strings(files)
		var n, likely, declared, pasted, available int
		var scores []int
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			img, format, _, err := security.ValidateUpload(bytes.NewReader(data), cfg.MaxUploadMB<<20)
			if err != nil {
				fmt.Printf("%s\t%s\tskipped: %v\n", label, filepath.Base(f), err)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), cfg.HardTimeout)
			res := deps.Pipeline.Run(ctx, pipeline.Request{InputType: models.InputFile, Image: img, ImageFormat: format})
			cancel()
			d := httpapi.BuildDossier(deps.Key, res, pipeline.Request{InputType: models.InputFile, Image: img, ImageFormat: format})
			v := d.Verdict

			ai := d.Pillars.AI
			prov := string(ai.Provenance.Kind)
			if ai.Provenance.Source != "" {
				prov += "/" + ai.Provenance.Source
			}
			det := "-"
			if ai.Detector != nil {
				det = fmt.Sprint(ai.Detector.Score)
			}
			text := -1
			for _, a := range d.Pillars.ELA.Anomalies {
				if a.Evidence != nil && *a.Evidence > text {
					text = *a.Evidence
				}
			}
			score := "-"
			if v.Score != nil {
				score = fmt.Sprint(*v.Score)
			}
			aiScore := "-"
			n++
			if ai.Available {
				available++
				aiScore = fmt.Sprint(ai.Score)
				scores = append(scores, ai.Score)
			}
			if slices.Contains(v.FloorsApplied, scoring.FloorAILikely) {
				likely++
			}
			if slices.Contains(v.FloorsApplied, scoring.FloorAIProvenance) || slices.Contains(v.FloorsApplied, scoring.FloorAIEdited) {
				declared++
			}
			if text >= pipeline.TextHotspotMin {
				pasted++
			}
			fmt.Printf("%s\t%s\t%s\t%.1f\t%s\t%s\t%d\t%s\t%s\t%s\n", label, filepath.Base(f), aiScore, ai.Reliability,
				prov, det, text, v.Status, score, strings.Join(v.FloorsApplied, ","))
			time.Sleep(250 * time.Millisecond) // stay well inside free-tier rate limits
		}
		fmt.Fprintf(os.Stderr, "%s: %d images, AI judged %d, median AI score %d, AI_LIKELY %d, declared AI %d, pasted text (>= %d) %d\n",
			label, n, available, median(scores), likely, declared, pipeline.TextHotspotMin, pasted)
	}
}

func median(v []int) int {
	if len(v) == 0 {
		return -1
	}
	s := slices.Clone(v)
	slices.Sort(s)
	return s[len(s)/2]
}
