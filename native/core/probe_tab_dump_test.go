package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestProbeTabDump(t *testing.T) {
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
	for _, tab := range []string{"8", "16"} {
		extra := url.Values{"tab_type": {tab}, "offset": {"0"}, "count": {"30"}}
		result, err := d.hongguoAppRequest(ctx, http.MethodGet, "/reading/bookapi/bookmall/tab/v1/", extra, nil)
		if err != nil {
			fmt.Printf("[tab=%s] ERR %v\n", tab, err)
			continue
		}
		raw, _ := json.Marshal(result)
		fmt.Printf("[tab=%s] len=%d\n", tab, len(raw))
		os.WriteFile("E:/dev/tab_"+tab+".json", raw, 0o644)
	}
}
