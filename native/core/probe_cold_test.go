package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProbeColdStart(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	for i := 0; i < 3; i++ {
		start := time.Now()
		page, err := engine.nativeCatalog(ctx, nativeInput{Source: "hongguo", Page: 1+i, Force: true})
		fmt.Printf("[cold %d] err=%v items=%d %.1fs\n", i, err, len(page.Items), time.Since(start).Seconds())
	}
}
