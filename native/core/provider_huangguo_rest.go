package core

// 黄果AI REST 客户端。
//
// 参考第三方客户端 HGhubX（https://github.com/zxsos/HGhubX，PolyForm 非商业许可）公开的
// 站点 REST 接口：游客可直接 GET，无需登录；返回信封为 {status:1,data:...}。
// 与网页抓取相比，接口分页/搜索/播放地址更稳定，且可随镜像线路切换。
// 站点域名经常更换，这里按「配置地址 → 内置镜像」顺序逐个尝试并记住可用线路。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	huangguoRESTMu     sync.Mutex
	huangguoRESTActive string
)

func (d *Downloader) huangguoRESTBases() []string {
	bases := make([]string, 0, len(huangguoAIBaseMirrors)+1)
	seen := map[string]bool{}
	add := func(raw string) {
		raw = strings.TrimRight(strings.TrimSpace(raw), "/")
		if raw == "" || seen[raw] {
			return
		}
		seen[raw] = true
		bases = append(bases, raw)
	}
	huangguoRESTMu.Lock()
	add(huangguoRESTActive)
	huangguoRESTMu.Unlock()
	add(d.cfg.HuangguoAIURL)
	for _, mirror := range huangguoAIBaseMirrors {
		add(mirror)
	}
	return bases
}

func (d *Downloader) huangguoRESTDo(ctx context.Context, method, path string, payload any) (map[string]any, error) {
	var body []byte
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = encoded
	}
	var lastErr error
	for _, base := range d.huangguoRESTBases() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		address := base + path
		request, err := http.NewRequestWithContext(ctx, method, address, nil)
		if err != nil {
			lastErr = err
			continue
		}
		if body != nil {
			request, err = http.NewRequestWithContext(ctx, method, address, strings.NewReader(string(body)))
			if err != nil {
				lastErr = err
				continue
			}
			request.Header.Set("Content-Type", "application/json; charset=utf-8")
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("Referer", base+"/")
		request.Header.Set("User-Agent", huangguoRESTUserAgent)
		client := &http.Client{Timeout: 15 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			lastErr = fmt.Errorf("%s 请求失败: %w", hostOf(base), err)
			continue
		}
		raw, readErr := io.ReadAll(io.LimitReader(response.Body, 4<<20))
		response.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if response.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("%s HTTP %d", hostOf(base), response.StatusCode)
			continue
		}
		var envelope map[string]any
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		if err := decoder.Decode(&envelope); err != nil {
			lastErr = fmt.Errorf("%s 返回格式异常", hostOf(base))
			continue
		}
		if fmt.Sprint(envelope["status"]) == "1" {
			huangguoRESTMu.Lock()
			huangguoRESTActive = base
			huangguoRESTMu.Unlock()
			data, _ := envelope["data"].(map[string]any)
			if data == nil {
				data = map[string]any{"list": envelope["data"]}
			}
			return data, nil
		}
		message := mapString(envelope, "msg", "message")
		if message == "" {
			message = "接口返回失败"
		}
		lastErr = fmt.Errorf("%s: %s", hostOf(base), message)
	}
	if lastErr == nil {
		lastErr = errors.New("黄果接口不可用")
	}
	return nil, lastErr
}

const huangguoRESTUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"

func hostOf(rawURL string) string {
	if parsed, err := url.Parse(rawURL); err == nil {
		return parsed.Host
	}
	return rawURL
}

// huangguoRESTCatalog 走 /api/videos 列表接口。category 为空时取全部。
func (d *Downloader) huangguoRESTCatalog(ctx context.Context, page int, category string) ([]Drama, bool, error) {
	if page < 1 {
		page = 1
	}
	size := 54
	path := fmt.Sprintf("/api/videos?page=%d&size=%d", page, size)
	if route := huangguoRESTCategoryRoute(category); route != "" {
		path = fmt.Sprintf("/api/videos/category/%s?page=%d&size=%d", url.PathEscape(route), page, size)
	}
	data, err := d.huangguoRESTDo(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, false, err
	}
	items := anyList(data["items"])
	if len(items) == 0 {
		items = anyList(data["list"])
	}
	dramas := make([]Drama, 0, len(items))
	for _, raw := range items {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id := mapString(row, "id", "video_id")
		if id == "" {
			continue
		}
		title := mapString(row, "title", "name")
		cover := mapString(row, "cover", "cover_url")
		total := int(mapNumber(row, "total_episodes", "episode_count"))
		episodes := int(mapNumber(row, "episode_count"))
		dramas = append(dramas, Drama{
			ID:             providerDramaID(sourceHuangguoAI, id),
			Source:         sourceHuangguoAI,
			SourceID:       id,
			Title:          title,
			Name:           title,
			Desc:           mapString(row, "description", "intro"),
			Intro:          mapString(row, "description", "intro"),
			Cover:          cover,
			CoverURL:       cover,
			CategoryName:   category,
			ChannelName:    "huangguoai.com",
			Score:          formatScore(mapFloat(row, "score")),
			TotalEpisode:   total,
			EpisodeCount:   episodes,
		})
	}
	hasMore := false
	if pagination, ok := data["pagination"].(map[string]any); ok {
		hasMore = int(mapNumber(pagination, "page")) < int(mapNumber(pagination, "pages"))
	}
	if !hasMore {
		hasMore = len(dramas) >= size
	}
	return dramas, hasMore, nil
}

func huangguoRESTCategoryRoute(category string) string {
	switch strings.TrimSpace(category) {
	case "", "全部", "推荐":
		return ""
	case "AI短剧", "ai-duanju":
		return "ai-duanju"
	case "AI漫剧", "ai-manju":
		return "ai-manju"
	default:
		return ""
	}
}

// huangguoRESTPlayback 走播放接口，返回可直连的播放地址。
// ep 为集数（从 1 开始）。优先新接口，失败回退旧接口。
func (d *Downloader) huangguoRESTPlayback(ctx context.Context, id string, ep int) (providerMedia, error) {
	data, err := d.huangguoRESTDo(ctx, http.MethodPost,
		fmt.Sprintf("/api/videos/%s/episodes/%d/playback", url.PathEscape(id), ep), map[string]any{})
	if err != nil || !huangguoRESTHasMedia(data) {
		legacy, legacyErr := d.huangguoRESTDo(ctx, http.MethodGet,
			fmt.Sprintf("/api/videos/%s/play?ep=%d&page_texts=0", url.PathEscape(id), ep), nil)
		if legacyErr == nil && huangguoRESTHasMedia(legacy) {
			data = legacy
		} else if err != nil {
			return providerMedia{}, err
		}
	}
	if !huangguoRESTHasMedia(data) {
		return providerMedia{}, errors.New("黄果接口未返回播放地址")
	}
	streams, _ := data["streams"].(map[string]any)
	address := mapString(streams, "main")
	if address == "" {
		address = mapString(data, "video_url", "videoUrl")
	}
	if address == "" {
		address = mapString(streams, "hevc")
	}
	if address == "" {
		return providerMedia{}, errors.New("黄果接口未返回播放地址")
	}
	duration := time.Duration(mapFloat(data, "duration") * float64(time.Second))
	media := providerMedia{URL: address, Referer: d.providerBaseURL(sourceHuangguoAI) + "/", Duration: duration}
	if hevc := mapString(streams, "hevc"); hevc != "" && hevc != address {
		variant := media
		variant.URL = hevc
		media.Variants = append(media.Variants, variant)
	}
	return media, nil
}

func huangguoRESTHasMedia(data map[string]any) bool {
	if len(data) == 0 {
		return false
	}
	streams, _ := data["streams"].(map[string]any)
	if mapString(streams, "main", "hevc") != "" {
		return true
	}
	return mapString(data, "video_url", "videoUrl") != ""
}

// huangguoRESTChapters 走详情接口拿分集列表。
func (d *Downloader) huangguoRESTChapters(ctx context.Context, id string) ([]Chapter, error) {
	data, err := d.huangguoRESTDo(ctx, http.MethodGet, "/api/videos/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	episodes := anyList(data["episodes"])
	if len(episodes) == 0 {
		return nil, errors.New("黄果详情未返回分集")
	}
	chapters := make([]Chapter, 0, len(episodes))
	for index, raw := range episodes {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		ep := int(mapNumber(row, "ep_num", "ep"))
		if ep <= 0 {
			ep = index + 1
		}
		title := mapString(row, "title")
		if title == "" {
			title = fmt.Sprintf("第%d集", ep)
		}
		chapters = append(chapters, Chapter{
			ID:             providerChapterID(sourceHuangguoAI, id, fmt.Sprint(ep)),
			Source:         sourceHuangguoAI,
			Title:          title,
			CurrentEpisode: rawEpisode(ep),
			VideoURL:       "",
		})
	}

	return chapters, nil
}

func mapNumber(row map[string]any, keys ...string) float64 {
	for _, key := range keys {
		switch value := row[key].(type) {
		case json.Number:
			if parsed, err := value.Float64(); err == nil {
				return parsed
			}
		case float64:
			return value
		case int:
			return float64(value)
		case string:
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			var parsed float64
			if _, err := fmt.Sscanf(trimmed, "%f", &parsed); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func mapFloat(row map[string]any, keys ...string) float64 {
	return mapNumber(row, keys...)
}

func formatScore(value float64) string {
	if value <= 0 {
		return ""
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", value), "0"), ".")
}
