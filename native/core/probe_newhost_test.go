package core

import (
	"context"
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

func TestProbeNewHosts(t *testing.T) {
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
	dev, iid, err := d.registerHongguoDevice(ctx)
	if err != nil {
		fmt.Println("REG ERR", err)
		return
	}
	vid := "7503780574541122622"
	targets := []struct {
		host, path string
		params     url.Values
	}{
		{"novel.snssdk.com", "/novel/player/video_model/v1/", url.Values{"video_id": {vid}}},
		{"novel.snssdk.com", "/novel/player/video_model/v1/", url.Values{"item_id": {vid}}},
		{"play.jinguoduanju.app", "/novel/player/video_model/v1/", url.Values{"video_id": {vid}}},
		{"play.jinguoduanju.app", "/novel/player/video_detail/v1/", url.Values{"video_id": {vid}}},
	}
	for _, tgt := range targets {
		now := time.Now()
		q := hongguoAppQuery(dev, iid)
		for k, v := range tgt.params {
			q.Set(k, v[0])
		}
		q.Set("_rticket", strconv.FormatInt(now.UnixMilli(), 10))
		address := "https://" + tgt.host + tgt.path + "?" + q.Encode()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		req.Header.Set("User-Agent", hongguoAppUserAgent)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-XS-From-Web", "0")
		req.Header.Set("Sdk-Version", "2")
		signHongguoPlayerRequest(req, nil, now, iid)
		resp, err := d.client.Do(req)
		if err != nil {
			fmt.Printf("[%s%s] NETERR %v\n", tgt.host, tgt.path, err)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		resp.Body.Close()
		fmt.Printf("[%s%s ?%v] HTTP %d len=%d head=%.200s\n", tgt.host, tgt.path, tgt.params, resp.StatusCode, len(body), strings.TrimSpace(string(body)))
		_ = client
	}
}
