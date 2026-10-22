package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProbeFingerprint(t *testing.T) {
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
	type fp struct {
		name, model, brand, osVer, osAPI, res string
	}
	variants := []fp{
		{"redmi-old", "25053RT47C", "Redmi", "16", "36", "1280*2772"},
		{"xiaomi14", "Xiaomi 14", "Xiaomi", "14", "34", "1200*2670"},
		{"m2102j2sc", "M2102J2SC", "Xiaomi", "13", "33", "1080*2400"},
		{"harmony", "PGT-AN10", "HUAWEI", "12", "31", "1220*2700"},
	}
	for _, v := range variants {
		dev, iid, err := d.registerHongguoDevice(ctx)
		if err != nil {
			fmt.Printf("[%s] REGERR %v\n", v.name, err)
			continue
		}
		ok, empty := 0, 0
		for i := 0; i < 2; i++ {
			now := time.Now()
			q := hongguoAppQuery(dev, iid)
			q.Set("device_model", v.model)
			q.Set("device_brand", v.brand)
			q.Set("os_version", v.osVer)
			q.Set("os_api", v.osAPI)
			q.Set("resolution", v.res)
			q.Set("_rticket", strconv.FormatInt(now.UnixMilli(), 10))
			address := "https://api5-normal-sinfonlineb.fqnovel.com/reading/distribution/category/landpage/v1/?" + q.Encode()
			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, address, strings.NewReader(string(payload)))
			req.Header.Set("User-Agent", hongguoAppUserAgent)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			req.Header.Set("X-XS-From-Web", "0")
			req.Header.Set("Sdk-Version", "2")
			signHongguoPlayerRequest(req, payload, now, iid)
			resp, err := d.client.Do(req)
			if err != nil {
				continue
			}
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			if len(body) > 0 {
				ok++
			} else {
				empty++
			}
			time.Sleep(time.Second)
		}
		fmt.Printf("[%s %s %s] ok=%d empty=%d\n", v.name, v.model, v.osVer, ok, empty)
	}
	_ = url.Values{}
}
