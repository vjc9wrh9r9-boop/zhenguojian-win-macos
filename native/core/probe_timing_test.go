package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestResolveTiming(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := engine.downloader
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cases := [][2]string{
		{"7503745441910017049", "7503780574541122622"}, // 普通老剧（网页有公开地址）
		{"7709114780467858456", "7709114791058492440"}, // 会员新剧（预期失败）
	}
	for _, c := range cases {
		started := time.Now()
		media, err := d.resolveHongguoMedia(ctx, Task{DramaID: "hongguo:" + c[0], Chapter: Chapter{VideoURL: "hongguo-cenc://" + c[1]}})
		fmt.Printf("[%s/%s] 用时=%.2fs err=%.60v url=%.60s\n", c[0][:4], c[1][:4], time.Since(started).Seconds(), err, media.URL)
	}
}
