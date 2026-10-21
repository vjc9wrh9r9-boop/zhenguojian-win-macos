package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestProbeSessionPaging(t *testing.T) {
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

	fetch := func(offset int, session string, reqType int) ([]string, string, error) {
		payload := map[string]any{
			"req_scene": "comic_series", "offset": offset, "limit": 54,
			"req_type": "only_content", "need_selector_panel": false, "client_req_type": reqType,
			"session_id": session, "filter_ids": "",
			"select_items": map[string]any{
				"genre": []string{"comic_series"}, "sort": []string{"online_time"}, "gender": []string{},
				"category_dim_theme": []string{}, "category_dim_role": []string{}, "category_dim_epoch": []string{},
				"online_time": []string{}, "creation_status": []string{},
			},
		}
		var result map[string]any
		var err error
		for attempt := 0; attempt < 6; attempt++ {
			result, err = d.hongguoAppRequest(ctx, http.MethodPost, "/reading/distribution/category/landpage/v/", nil, payload)
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			return nil, "", err
		}
		data := nestedMap(result, "data")
		var ids []string
		for _, row := range anyList(data["video_data"]) {
			m := row.(map[string]any)
			id := mapString(m, "book_id", "id", "video_id")
			if id != "" {
				ids = append(ids, id)
			}
		}
		return ids, mapString(data, "session_id"), nil
	}

	ids1, s1, err := fetch(0, "", 3)
	fmt.Printf("p1: n=%d session=%v err=%v first=%v\n", len(ids1), len(s1) > 0, err, first(ids1))
	ids2, s2, err := fetch(54, s1, 2)
	fmt.Printf("p2(session): n=%d err=%v same_as_p1=%v first=%v\n", len(ids2), err, same(ids1, ids2), first(ids2))
	ids3, _, err := fetch(54, "", 3)
	fmt.Printf("p3(no session): n=%d err=%v same_as_p1=%v first=%v\n", len(ids3), err, same(ids1, ids3), first(ids3))
	_ = s2
}

func first(ids []string) string {
	if len(ids) == 0 {
		return "-"
	}
	return ids[0]
}

func same(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return a[0] == b[0]
}
