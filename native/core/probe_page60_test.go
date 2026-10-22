package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProbePage60(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	page, err := engine.nativeCatalog(ctx, nativeInput{Source: "hongguo", Page: 60, Force: true})
	fmt.Printf("PAGE60 err=%v items=%d hasMore=%v warning=%q\n", err, len(page.Items), page.HasMore, page.Warning)
}
