package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProbeWebPlay(t *testing.T) {
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
	for _, pair := range [][2]string{
		{"7503745441910017049", "7503780574541122622"},
		{"7503745441910017049", "7503780574541122623"},
	} {
		media, err := d.resolveHongguoWebMedia(ctx, pair[0], pair[1])
		fmt.Printf("[web %s] err=%v url=%.80s\n", pair[1], err, media.URL)
	}
	// 完整解析链
	task := Task{DramaID: "hongguo:7503745441910017049", Chapter: Chapter{VideoURL: "hongguo-cenc://7503780574541122622"}}
	media, err := d.resolveHongguoMedia(ctx, task)
	fmt.Printf("[chain] err=%v url=%.90s\n", err, media.URL)
}
