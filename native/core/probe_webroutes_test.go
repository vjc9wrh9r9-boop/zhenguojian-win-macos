package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProbeWebRoutes(t *testing.T) {
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
	for _, route := range []string{"real-drama", "comic-drama", "comic", "ai-drama"} {
		items, pages, err := d.fetchHongguoCategoryPage(ctx, route+"?page=1", route)
		first := ""
		if len(items) > 0 {
			first = items[0].Title
		}
		fmt.Printf("[%s] err=%v items=%d pages=%d first=%q\n", route, err, len(items), pages, first)
	}
}
