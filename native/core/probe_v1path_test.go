package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestProbeV1Path(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := engine.downloader
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	payload := map[string]any{
		"req_scene": "default", "offset": 0, "limit": 54,
		"req_type": "only_content", "need_selector_panel": false, "client_req_type": 3,
		"session_id": "", "filter_ids": "",
		"select_items": map[string]any{
			"genre": []string{"short_play"}, "sort": []string{"online_time"}, "gender": []string{},
			"category_dim_theme": []string{}, "category_dim_role": []string{}, "category_dim_epoch": []string{},
			"online_time": []string{}, "creation_status": []string{},
		},
	}
	for _, path := range []string{
		"/reading/distribution/category/landpage/v/",
		"/reading/distribution/category/landpage/v1/",
		"/reading/bookapi/bookmall/tab/v1/",
	} {
		for attempt := 0; attempt < 3; attempt++ {
			result, err := d.hongguoAppRequest(ctx, http.MethodPost, path, nil, payload)
			if err == nil {
				data := nestedMap(result, "data")
				rows := anyList(data["video_data"])
				fmt.Printf("[%s] OK rows=%d hasMore=%v\n", path, len(rows), data["has_more"])
				break
			}
			if attempt == 2 {
				fmt.Printf("[%s] ERR %v\n", path, err)
			}
			time.Sleep(time.Second)
		}
	}
}
