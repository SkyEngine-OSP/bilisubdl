package bilibili

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/K0ng2/bilisubdl/utils"
)

// api url examples
/*
title
https://api.bilibili.tv/intl/gateway/web/v2/ogv/play/season_info?season_id=1049041

episode list
https://api.bilibili.tv/intl/gateway/web/v2/ogv/play/episodes?season_id=1049041

episode
https://api.bilibili.tv/intl/gateway/web/v2/subtitle?s_locale&episode_id=368729
*/

const bilibiliAPI string = "https://api.bilibili.tv/intl/gateway"
const bilibiliInfoAPI string = bilibiliAPI + "/web/v2/ogv/play/"

// const bilibiliEpisodeAPI string = bilibiliAPI + "/subtitle?s_locale&episode_id="
const bilibiliEpisodeAPI string = bilibiliAPI + "/m/subtitle"

func GetInfo(id string) (*Info, error) {
	var info = new(Info)
	query := map[string]string{
		"s_locale":  "en_US",
		"season_id": id,
	}
	resp, err := utils.Request(bilibiliInfoAPI+"season_info", query)
	if err != nil {
		return nil, err
	}
	if resp.Json(info); err != nil {
		return nil, err
	}
	if info.Code != 0 {
		return nil, fmt.Errorf("api response code %d: %s", info.Code, info.Message)
	}
	return info, nil
}

func GetEpisodes(id string) (*Episodes, error) {
	var epList = new(Episodes)
	query := map[string]string{
		"s_locale":  "en_US",
		"season_id": id,
	}
	resp, err := utils.Request(bilibiliInfoAPI+"episodes", query)
	if err != nil {
		return nil, err
	}
	if resp.Json(epList); err != nil {
		return nil, err
	}
	if epList.Code != 0 {
		return nil, fmt.Errorf("api response code %d: %s", epList.Code, epList.Message)
	}
	return epList, nil
}

func GetEpisode(id string) (*Episode, error) {
	var ep = new(Episode)
	query := map[string]string{
		"s_locale":   "en_US",
		"ep_id": id,
	}
	resp, err := utils.Request(bilibiliEpisodeAPI, query)
	if err != nil {
		return nil, err
	}
	if resp.Json(ep); err != nil {
		return nil, err
	}
	if ep.Code != 0 {
		return nil, fmt.Errorf("api response code %d: %s", ep.Code, ep.Message)
	}
	return ep, nil
}

func (s *Episode) Subtitle(language string) ([]byte, string, error) {
	var index int = -1
	for i, s := range s.Data.Subtitles {
		if s.Key == language {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, "", fmt.Errorf("language \"%s\" not found", language)
	}
	if s.Data.Subtitles[index].IsMachine {
		fmt.Println("Warning machine translation")
	}

	resp, err := utils.Request(s.Data.Subtitles[index].URL, nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	fileType := filepath.Ext(resp.Request.URL.Path)
	switch fileType {
	case ".json":
		var subJson = new(Subtitle)
		err := resp.Json(subJson)
		if err != nil {
			return nil, "", err
		}
		return []byte(jsonToSRT(subJson)), ".srt", nil
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}
		return body, fileType, nil
	}
}

func jsonToSRT(subJson *Subtitle) string {
	var sub []string
	var content string
	for i, s := range subJson.Body {
		content = s.Content
		if s.Location != 2 {
			content = fmt.Sprintf("{\\an%d}%s", s.Location, content)
		}
		sub = append(sub, fmt.Sprintf("%d\n%s --> %s\n%s", i+1, utils.SecondToTime(s.From), utils.SecondToTime(s.To), content))
	}
	return strings.Join(sub, "\n\n") + "\n"
}

func GetTimeline() (*Timeline, error) {
	var timeline = new(Timeline)
	resp, err := utils.Request(bilibiliAPI+"/web/v2/ogv/timeline", nil)
	if err != nil {
		return nil, err
	}
	if resp.Json(timeline); err != nil {
		return nil, err
	}
	if timeline.Code != 0 {
		return nil, fmt.Errorf("api response code %d: %s", timeline.Code, timeline.Message)
	}
	return timeline, nil
}

func GetSearch(s string) (*Search, error) {
	var search = new(Search)
	query := map[string]string{
		"keyword":  s,
		"platform": "web",
		"pn":       "1",
		"ps":       "10",
		"s_locale": "en_US",
	}
	resp, err := utils.Request(bilibiliAPI+"/web/v2/search", query)
	if err != nil {
		return nil, err
	}

	if resp.Json(search); err != nil {
		return nil, err
	}
	if search.Code != 0 {
		return nil, fmt.Errorf("api response code %d: %s", search.Code, search.Message)
	}
	return search, nil
}
