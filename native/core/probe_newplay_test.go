package core

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestProbeNewPlay(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := engine.downloader
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	vid := "7503780574541122622"
	sid := "7503745441910017049"
	cases := []struct {
		path  string
		extra url.Values
	}{
		{"/novel/player/video_model/v1/", url.Values{"video_id": {vid}}},
		{"/novel/player/video_model/v1/", url.Values{"item_id": {vid}}},
		{"/novel/player/video_model/v1/", url.Values{"vid": {vid}}},
		{"/novel/player/video_model/v1/", url.Values{"video_id": {vid}, "series_id": {sid}}},
		{"/novel/player/video_detail/v1/", url.Values{"video_id": {vid}}},
		{"/novel/player/video_model/v1/", url.Values{"episode_id": {vid}, "series_id": {sid}}},
	}
	for _, c := range cases {
		res, err := d.hongguoAppRequest(ctx, http.MethodGet, c.path, c.extra, nil)
		if err != nil {
			fmt.Printf("[%s ?%v] ERR %v\n", c.path, c.extra, err)
			continue
		}
		fmt.Printf("[%s ?%v] OK keys=%v\n", c.path, c.extra, keysOf(res))
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, 6)
	for k := range m {
		out = append(out, k)
		if len(out) >= 6 {
			break
		}
	}
	return out
}
