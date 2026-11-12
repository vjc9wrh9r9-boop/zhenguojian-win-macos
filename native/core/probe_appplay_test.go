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

func TestProbeAppPlay(t *testing.T) {
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
	dev, iid, err := d.registerHongguoDevice(ctx)
	if err != nil {
		fmt.Println("REG ERR", err)
		return
	}
	vid := "7503780574541122622"
	hosts := []string{"api5-normal-sinfonlineb.fqnovel.com", "api5-normal-sinfonlinea.fqnovel.com", "api3-normal-sinfonlinea.fqnovel.com"}
	paths := []string{"/novel/player/video_model/v1/", "/reading/bookapi/multi_video_model/v1/"}
	bodies := []map[string]any{{"video_id": vid}, {"item_id": vid}, {"video_ids": []string{vid}}}
	for _, host := range hosts[:2] {
		for _, path := range paths {
			for _, bodyMap := range bodies[:1] {
				body, _ := json.Marshal(bodyMap)
				now := time.Now()
				q := hongguoAppQuery(dev, iid)
				q.Set("_rticket", strconv.FormatInt(now.UnixMilli(), 10))
				address := "https://" + host + path + "?" + q.Encode()
				req, _ := http.NewRequestWithContext(ctx, http.MethodPost, address, strings.NewReader(string(body)))
				req.Header.Set("User-Agent", hongguoAppUserAgent)
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				req.Header.Set("Accept", "application/json")
				req.Header.Set("X-XS-From-Web", "0")
				req.Header.Set("Sdk-Version", "2")
				signHongguoPlayerRequest(req, body, now, iid)
				resp, err := d.client.Do(req)
				if err != nil {
					fmt.Printf("[%s%s] NETERR %v\n", host, path, err)
					continue
				}
				raw, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
				resp.Body.Close()
				fmt.Printf("[%s%s] HTTP %d len=%d %.120s\n", host, path, resp.StatusCode, len(raw), strings.TrimSpace(string(raw)))
			}
		}
	}
}
