package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestProbeNow(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := engine.downloader
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	payload := map[string]any{
		"req_scene": "comic_series", "offset": 0, "limit": 54,
		"req_type": "only_content", "need_selector_panel": false, "client_req_type": 3,
		"session_id": "", "filter_ids": "",
		"select_items": map[string]any{
			"genre": []string{"comic_series"}, "sort": []string{"online_time"}, "gender": []string{},
			"category_dim_theme": []string{}, "category_dim_role": []string{}, "category_dim_epoch": []string{},
			"online_time": []string{}, "creation_status": []string{},
		},
	}
	for i := 0; i < 5; i++ {
		start := time.Now()
		_, err := d.hongguoAppRequest(ctx, http.MethodPost, "/reading/distribution/category/landpage/v/", nil, payload)
		fmt.Printf("[probe %d] err=%v %.1fs\n", i, err, time.Since(start).Seconds())
		time.Sleep(8 * time.Second)
	}
}
