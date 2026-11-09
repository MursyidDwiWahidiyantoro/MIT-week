package pipeline

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/scoring"
	"github.com/amba/amba/backend/internal/video"
)

const tiktokURL = "https://www.tiktok.com/@akun/video/7300000000000000000"

// fakeVideo stands in for yt-dlp and ffmpeg.
type fakeVideo struct {
	clip  *video.Clip
	err   error
	calls int32
}

func (f *fakeVideo) Enabled() bool { return true }

func (f *fakeVideo) Fetch(context.Context, string) (*video.Clip, error) {
	atomic.AddInt32(&f.calls, 1)
	return f.clip, f.err
}

func testClip(t *testing.T) *video.Clip {
	return &video.Clip{
		Platform: "tiktok", Title: "Banjir besar hari ini", Description: "Segera sebarkan", UploadDate: "20261109",
		DurationSec: 31.44,
		Frames: []video.Frame{
			{AtSec: 1.52, JPEG: testJPEG(t, 1)}, {AtSec: 15, JPEG: testJPEG(t, 2)}, {AtSec: 29.9, JPEG: testJPEG(t, 3)},
		},
	}
}

func ogPage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`<html><head><meta property="og:title" content="Banjir besar hari ini"></head></html>`))
}

func TestVideoLinkIsJudgedOnItsFrames(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, ogPage)
	defer srv.Close()
	fv := &fakeVideo{clip: testClip(t)}
	p.Video = fv

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	v := res.Video
	if v == nil || v.Status != models.VideoOK || v.Reason != "" || v.Platform != "tiktok" ||
		v.UploadDate != "2026-11-09" || v.DurationSec != 31.4 {
		t.Fatalf("video = %+v", v)
	}
	if len(v.Frames) != 3 || v.Frames[0].AtSec != 1.5 || !strings.HasPrefix(v.Frames[2].Thumb, "data:image/jpeg;base64,") {
		t.Fatalf("frames = %+v", v.Frames)
	}
	if v.Summary == "" || !res.Pillars.AI.Available {
		t.Fatalf("summary %q, AI pillar %+v", v.Summary, res.Pillars.AI)
	}
	// Whatever the mock found, the floors must follow the video's findings.
	floors := Verdict(res, false).FloorsApplied
	if (res.Pillars.AI.Score >= scoring.AILikelyMin) != slices.Contains(floors, scoring.FloorAILikely) {
		t.Fatalf("AI score %d, floors %v", res.Pillars.AI.Score, floors)
	}
	if (v.Context.Score >= scoring.VideoSignalMin) != slices.Contains(floors, scoring.FloorVideoContextMismatch) {
		t.Fatalf("context %d, floors %v", v.Context.Score, floors)
	}
}

func TestPlatformAILabelIsProvenance(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, ogPage)
	defer srv.Close()
	clip := testClip(t)
	clip.AILabel = true
	p.Video = &fakeVideo{clip: clip}

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	if !res.Video.PlatformAILabel || res.Pillars.AI.Provenance.Kind != models.ProvenanceAIGenerated {
		t.Fatalf("video %+v, provenance %+v", res.Video, res.Pillars.AI.Provenance)
	}
	if v := Verdict(res, false); !slices.Contains(v.FloorsApplied, scoring.FloorAIProvenance) || v.Status != models.StatusTampered {
		t.Fatalf("verdict = %+v", v)
	}
}

func TestPrivateVideoFallsBackToThePreview(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, ogPage)
	defer srv.Close()
	fv := &fakeVideo{err: video.ErrPrivate}
	p.Video = fv

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	if res.Video == nil || res.Video.Status != models.VideoUnavailable || res.Video.Reason != "private" || len(res.Video.Frames) != 0 {
		t.Fatalf("video = %+v", res.Video)
	}
	if slices.Contains(Verdict(res, false).FloorsApplied, scoring.FloorAILikely) {
		t.Fatal("an unavailable video must not be judged like one")
	}
}

func TestTransientVideoFailureIsNotCached(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, ogPage)
	defer srv.Close()
	fv := &fakeVideo{err: video.ErrTimeout}
	p.Video = fv

	for i := 0; i < 2; i++ {
		res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
		if res.Video == nil || res.Video.Reason != "timeout" || !res.Degraded {
			t.Fatalf("run %d: video %+v degraded %v", i, res.Video, res.Degraded)
		}
	}
	if n := atomic.LoadInt32(&fv.calls); n != 2 {
		t.Fatalf("fetched %d times, want 2 (a failed download is not cached)", n)
	}
}

func TestOtherLinksAreNotDownloaded(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, ogPage)
	defer srv.Close()
	fv := &fakeVideo{clip: testClip(t)}
	p.Video = fv

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: "https://berita.example/banjir"})
	if res.Video != nil || atomic.LoadInt32(&fv.calls) != 0 {
		t.Fatalf("video = %+v, fetches = %d", res.Video, fv.calls)
	}
}

func TestVideoReasonCodes(t *testing.T) {
	for _, tc := range []struct {
		err       error
		reason    string
		transient bool
	}{
		{video.ErrUnsupported, "unsupported", false},
		{video.ErrPrivate, "private", false},
		{video.ErrTooLong, "too_long", false},
		{video.ErrTooBig, "too_big", false},
		{video.ErrTimeout, "timeout", true},
		{context.DeadlineExceeded, "timeout", true},
		{video.ErrNoTools, "no_tools", true},
		{errors.New("anything else"), "download_failed", true},
	} {
		if r, tr := videoReason(tc.err); r != tc.reason || tr != tc.transient {
			t.Errorf("%v: got %q/%v, want %q/%v", tc.err, r, tr, tc.reason, tc.transient)
		}
	}
	if isoDate("20261109") != "2026-11-09" || isoDate("2026110") != "" || isoDate("2026-11-") != "" {
		t.Fatal("isoDate")
	}
}
