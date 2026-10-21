package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProbeRegister(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := engine.downloader
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	deviceID, installID, err := d.registerHongguoDevice(ctx)
	fmt.Printf("REGISTER dev=%s iid=%s err=%v\n", deviceID, installID, err)
	if err == nil {
		if err := d.ensureHongguoDevice(ctx, ""); err != nil {
			fmt.Printf("ENSURE err=%v\n", err)
		} else {
			client := d.hongguoClient()
			fmt.Printf("STATE dev=%s registered=%v\n", client.state.DeviceID, client.state.Registered)
		}
	}
}
