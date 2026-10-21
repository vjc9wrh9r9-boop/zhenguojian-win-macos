package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProbeWebCategory(t *testing.T) {
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
	items, pages, err := d.fetchHongguoCategoryPage(ctx, "real-drama?page=1", "真人剧")
	fmt.Printf("WEBCAT err=%v items=%d pages=%d\n", err, len(items), pages)
}

func TestProbeWebCategoryPages(t *testing.T) {
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
	for _, page := range []string{"2", "3", "5", "10"} {
		items, pages, err := d.fetchHongguoCategoryPage(ctx, "real-drama?page="+page, "真人剧")
		fmt.Printf("[page=%s] err=%v items=%d pages=%d\n", page, err, len(items), pages)
	}
}
