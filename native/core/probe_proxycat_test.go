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

func TestProbeCategoryViaProxy(t *testing.T) {
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
	client := d.hongguoClient()
	// 先注册设备
	dev, iid, err := d.registerHongguoDevice(ctx)
	if err != nil {
		fmt.Println("REGISTER err", err)
		return
	}
	fmt.Println("dev", dev)

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
	do := func(label string, transport *http.Transport) {
		now := time.Now()
		q := hongguoAppQuery(dev, iid)
		q.Set("_rticket", strconv.FormatInt(now.UnixMilli(), 10))
		address := strings.TrimRight(client.baseURL, "/") + "/reading/distribution/category/landpage/v/?" + q.Encode()
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, address, strings.NewReader(string(payload)))
		req.Header.Set("User-Agent", hongguoAppUserAgent)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("X-XS-From-Web", "0")
		req.Header.Set("Sdk-Version", "2")
		signHongguoPlayerRequest(req, payload, now, iid)
		httpClient := &http.Client{Transport: transport, Timeout: 20 * time.Second}
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Printf("[%s] NETERR %v\n", label, err)
			return
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		resp.Body.Close()
		fmt.Printf("[%s] HTTP %d len=%d head=%.120s\n", label, resp.StatusCode, len(body), strings.TrimSpace(string(body)))
	}
	do("direct", http.DefaultTransport.(*http.Transport).Clone())
	proxyURL, _ := url.Parse("http://127.0.0.1:7897")
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = http.ProxyURL(proxyURL)
	do("proxy", tr)
}
