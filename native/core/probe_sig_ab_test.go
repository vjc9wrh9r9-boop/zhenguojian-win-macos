package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProbeSigAB(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := engine.downloader
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	client := d.hongguoClient()
	dev, iid, err := d.registerHongguoDevice(ctx)
	if err != nil {
		fmt.Println("REGISTER err", err)
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"req_scene": "default", "offset": 0, "limit": 54,
		"req_type": "only_content", "need_selector_panel": false, "client_req_type": 3,
		"session_id": "", "filter_ids": "",
		"select_items": map[string]any{
			"genre": []string{"short_play"}, "sort": []string{"online_time"}, "gender": []string{},
			"category_dim_theme": []string{}, "category_dim_role": []string{}, "category_dim_epoch": []string{},
			"online_time": []string{}, "creation_status": []string{},
		},
	})
	modes := []string{"full", "no-stub", "old-sign", "no-argus"}
	for _, mode := range modes {
		ok, empty := 0, 0
		for i := 0; i < 3; i++ {
			now := time.Now()
			q := hongguoAppQuery(dev, iid)
			q.Set("_rticket", strconv.FormatInt(now.UnixMilli(), 10))
			address := strings.TrimRight(client.baseURL, "/") + "/reading/distribution/category/landpage/v1/?" + q.Encode()
			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, address, strings.NewReader(string(payload)))
			req.Header.Set("User-Agent", hongguoAppUserAgent)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			req.Header.Set("X-XS-From-Web", "0")
			req.Header.Set("Sdk-Version", "2")
			switch mode {
			case "full":
				signHongguoPlayerRequest(req, payload, now, iid)
			case "no-stub":
				signHongguoPlayerRequest(req, payload, now, iid)
				req.Header.Del("X-SS-STUB")
			case "old-sign":
				signHongguoRequest(req, payload, now)
			case "no-argus":
				signHongguoPlayerRequest(req, payload, now, iid)
				req.Header.Del("X-Argus")
				req.Header.Del("X-Ladon")
			}
			resp, err := d.client.Do(req)
			if err != nil {
				continue
			}
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			resp.Body.Close()
			if len(body) > 0 {
				ok++
			} else {
				empty++
			}
			time.Sleep(700 * time.Millisecond)
		}
		fmt.Printf("[%s] ok=%d empty=%d\n", mode, ok, empty)
	}
}
