package core

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const hongguoBookmallTabPath = "/reading/bookapi/bookmall/tab/v1/"

var hongguoRecommendationFeeds = []struct {
	key       string
	tabType   string
	bottomTab string
}{
	{key: "home_feed", tabType: "16", bottomTab: "7"},
	{key: "find_drama", tabType: "8"},
}

type hongguoRecommendationQuery struct {
	Genre     string   `json:"genre"`
	Offset    int      `json:"offset"`
	SessionID string   `json:"sessionId"`
	Seen      []string `json:"seen"`
}

type hongguoRecommendationPage struct {
	Dramas     []Drama `json:"data"`
	NextOffset int     `json:"nextOffset"`
	SessionID  string  `json:"sessionId"`
	HasMore    bool    `json:"hasMore"`
	Saved      bool    `json:"saved"`
}

func (query hongguoRecommendationQuery) validate() error {
	known := false
	for _, genre := range hongguoAppGenres {
		known = known || query.Genre == genre.key
	}
	for _, feed := range hongguoRecommendationFeeds {
		known = known || query.Genre == feed.key
	}
	if !known || query.Offset < 0 || query.Offset > 1_000_000 || len(query.SessionID) > 4096 || strings.ContainsAny(query.SessionID, "\r\n\x00") || len(query.Seen) > 540 {
		return errors.New("推荐分类或分页参数无效，请重新获取")
	}
	for _, id := range query.Seen {
		if !hongguoNumericID.MatchString(strings.TrimPrefix(id, "hongguo:")) {
			return errors.New("推荐分页含无效剧集 ID")
		}
	}
	return nil
}

func (downloader *Downloader) fetchHongguoRecommendations(ctx context.Context, query hongguoRecommendationQuery) (hongguoRecommendationPage, error) {
	page := hongguoRecommendationPage{Dramas: []Drama{}}
	if err := query.validate(); err != nil {
		return page, err
	}
	for _, feed := range hongguoRecommendationFeeds {
		if feed.key == query.Genre {
			return downloader.fetchHongguoFeedRecommendations(ctx, query, feed.tabType, feed.bottomTab)
		}
	}
	scene, category := "", ""
	for _, genre := range hongguoAppGenres {
		if genre.key == query.Genre {
			scene, category = genre.scene, genre.name
		}
	}
	seen := make(map[string]bool, len(query.Seen))
	filterIDs := make([]string, 0, len(query.Seen))
	for _, id := range query.Seen {
		id = strings.TrimPrefix(id, "hongguo:")
		if !seen[id] {
			filterIDs = append(filterIDs, id)
			seen[id] = true
		}
	}
	payload := map[string]any{
		"req_scene": scene, "offset": query.Offset, "limit": 54,
		"req_type": "only_content", "need_selector_panel": false, "client_req_type": 3,
		"session_id": query.SessionID, "filter_ids": strings.Join(filterIDs, ","),
		"select_items": map[string]any{
			"genre": []string{query.Genre}, "sort": []string{}, "gender": []string{},
			"category_dim_theme": []string{}, "category_dim_role": []string{}, "category_dim_epoch": []string{},
			"online_time": []string{}, "creation_status": []string{},
		},
	}
	if query.Offset > 0 {
		payload["client_req_type"] = 2
	}
	result, err := downloader.hongguoAppRequest(ctx, http.MethodPost, "/reading/distribution/category/landpage/v/", nil, payload)
	if err != nil {
		return page, err
	}
	data := nestedMap(result, "data")
	rows, valid := data["video_data"].([]any)
	hasMore, paginationOK := data["has_more"].(bool)
	next, parseErr := strconv.Atoi(mapString(data, "next_offset"))
	if !valid || !paginationOK || hasMore && parseErr != nil {
		return page, errors.New("红果推荐数据格式异常，已保留上次位置")
	}
	validRows := 0
	for _, row := range rows {
		drama := hongguoDramaFromAny(row, category)
		if drama.ID != "" {
			validRows++
		}
		id := strings.TrimPrefix(drama.ID, "hongguo:")
		if drama.ID == "" || seen[id] {
			continue
		}
		seen[id] = true
		page.Dramas = append(page.Dramas, drama)
	}
	if len(rows) > 0 && validRows == 0 {
		return hongguoRecommendationPage{}, errors.New("红果推荐未返回可识别的剧集")
	}
	if hasMore && (next <= query.Offset || next > 1_000_000 || len(page.Dramas) == 0) {
		return hongguoRecommendationPage{}, errors.New("红果推荐分页未前进，可重试或重新获取")
	}
	page.NextOffset, page.HasMore, page.SessionID = next, hasMore, mapString(data, "session_id")
	if !hasMore && (next < query.Offset || next > 1_000_000) {
		page.NextOffset = query.Offset
	}
	if len(page.SessionID) > 4096 || strings.ContainsAny(page.SessionID, "\r\n\x00") {
		return hongguoRecommendationPage{}, errors.New("红果推荐分页标记无效")
	}
	return page, ctx.Err()
}

func (downloader *Downloader) fetchHongguoFeedRecommendations(ctx context.Context, query hongguoRecommendationQuery, tabType, bottomTab string) (hongguoRecommendationPage, error) {
	seen := make(map[string]bool, len(query.Seen))
	for _, id := range query.Seen {
		seen[strings.TrimPrefix(id, "hongguo:")] = true
	}
	page := hongguoRecommendationPage{Dramas: []Drama{}, NextOffset: query.Offset, SessionID: query.SessionID, HasMore: true}
	validRows := 0
	for attempt := 0; attempt < 3 && len(page.Dramas) == 0; attempt++ {
		extra := url.Values{"tab_type": {tabType}, "offset": {strconv.Itoa(page.NextOffset)}, "client_req_type": {"1"}}
		if bottomTab != "" {
			extra.Set("bottom_tab_type", bottomTab)
			extra.Set("landing_bottom_tab_type", bottomTab)
		}
		if page.SessionID != "" {
			extra.Set("session_id", page.SessionID)
			extra.Set("client_req_type", "2")
		}
		result, err := downloader.hongguoAppRequest(ctx, http.MethodGet, hongguoBookmallTabPath, extra, nil)
		if err != nil {
			return hongguoRecommendationPage{}, err
		}
		var tab map[string]any
		for _, item := range anyList(nestedMap(result, "data")["tab_item"]) {
			if row, _ := item.(map[string]any); mapString(row, "tab_type") == tabType {
				tab = row
			}
		}
		if tab == nil {
			return hongguoRecommendationPage{}, errors.New("红果推荐数据格式异常，已保留上次位置")
		}
		if page.SessionID == "" {
			page.SessionID = mapString(tab, "session_id")
		}
		page.NextOffset++
		for _, cell := range anyList(tab["cell_data"]) {
			cellMap, _ := cell.(map[string]any)
			for _, card := range anyList(cellMap["cell_data"]) {
				cardMap, _ := card.(map[string]any)
				for _, video := range anyList(cardMap["video_data"]) {
					row, _ := video.(map[string]any)
					if detail, _ := row["video_detail"].(map[string]any); len(detail) > 0 {
						row = detail
					}
					drama := hongguoDramaFromAny(row, "")
					if drama.ID == "" {
						continue
					}
					validRows++
					if id := strings.TrimPrefix(drama.ID, "hongguo:"); !seen[id] {
						seen[id] = true
						page.Dramas = append(page.Dramas, drama)
					}
				}
			}
		}
	}
	if validRows == 0 {
		return hongguoRecommendationPage{}, errors.New("红果推荐未返回可识别的剧集")
	}
	if len(page.Dramas) == 0 || page.NextOffset > 1_000_000 {
		return hongguoRecommendationPage{}, errors.New("红果推荐分页未前进，可重试或重新获取")
	}
	if len(page.SessionID) > 4096 || strings.ContainsAny(page.SessionID, "\r\n\x00") {
		return hongguoRecommendationPage{}, errors.New("红果推荐分页标记无效")
	}
	return page, ctx.Err()
}
