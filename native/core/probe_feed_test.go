package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestProbeFeedNow(t *testing.T) {
	if os.Getenv("HG_PROBE") == "" {
		t.Skip("probe disabled")
	}
	engine, err := newNativeEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := engine.downloader
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()

	// 主推送流（不走分类接口）
	page, err := engine.nativeCatalog(ctx, nativeInput{Source: "hongguo", Page: 1, Force: true})
	fmt.Printf("FEED err=%v items=%d warning=%q\n", err, len(page.Items), page.Warning)

	// 详情
	_, chapters, err := d.GetHuangguoChapters(ctx, "hongguo", "7503745441910017049")
	fmt.Printf("DETAIL err=%v chapters=%d\n", err, len(chapters))

	// 搜索
	entry, err := d.searchHongguoDramasProgress(ctx, "霸总", nil)
	fmt.Printf("SEARCH err=%v items=%d\n", err, len(entry.Dramas))

	// 播放（multi_video_model 协议）
	media, err := d.resolveHongguoMultiMedia(ctx, "7503780574541122622")
	fmt.Printf("PLAYBACK err=%v ok=%v\n", err, media.CENCKey != nil)

	// 分类（App landpage）
	payload := map[string]any{
		"req_scene": "default", "offset": 0, "limit": 54,
		"req_type": "only_content", "need_selector_panel": false, "client_req_type": 3,
		"session_id": "", "filter_ids": "",
		"select_items": map[string]any{
			"genre": []string{"short_play"}, "sort": []string{"online_time"}, "gender": []string{},
			"category_dim_theme": []string{}, "category_dim_role": []string{}, "category_dim_epoch": []string{},
			"online_time": []string{}, "creation_status": []string{},
		},
	}
	_, err = d.hongguoAppRequest(ctx, http.MethodPost, "/reading/distribution/category/landpage/v/", nil, payload)
	fmt.Printf("CATEGORY err=%v\n", err)
}
