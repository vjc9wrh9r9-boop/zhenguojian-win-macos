package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestProbeBookmall(t *testing.T) {
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
	for _, tt := range []string{"16", "8"} {
		for _, offset := range []int{0, 54} {
			result, err := d.hongguoAppRequest(ctx, http.MethodGet, "/reading/bookapi/bookmall/tab/v1/", nil, nil)
			_ = result
			_ = err
			// 直接构造 GET 更可控：带上 tab_type/offset
			_ = tt
			_ = offset
			break
		}
		break
	}
	// 直接 GET 原始接口看结构
	for _, tab := range []string{"16", "8"} {
		extra := map[string]string{"tab_type": tab, "offset": "0", "count": "54"}
		q := hongguoAppQuery("", "")
		for k, v := range extra {
			q.Set(k, v)
		}
		address := "https://api5-normal-sinfonlineb.fqnovel.com/reading/bookapi/bookmall/tab/v1/?" + q.Encode()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		req.Header.Set("User-Agent", hongguoAppUserAgent)
		req.Header.Set("Accept", "application/json")
		resp, err := d.client.Do(req)
		if err != nil {
			fmt.Printf("[tab=%s] NETERR %v\n", tab, err)
			continue
		}
		body := make([]byte, 2048)
		n, _ := resp.Body.Read(body)
		resp.Body.Close()
		fmt.Printf("[tab=%s] HTTP %d len=%d head=%.200s\n", tab, resp.StatusCode, n, string(body[:200]))
	}
}

