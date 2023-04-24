package bilibili

import (
	"fmt"
	"io"
	"strings"

	"github.com/K0ng2/bilisubdl/utils"
	"golang.org/x/exp/slices"
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

const (
	BilibiliAPI            string = "https://api.bilibili.tv/intl/gateway"
	BilibiliInfoAPI        string = BilibiliAPI + "/web/v2/ogv/play/"
	BilibiliSeasonInfoAPI  string = BilibiliInfoAPI + "season_info"
	BilibiliEpisodeInfoAPI string = BilibiliInfoAPI + "episodes"
	BilibiliSubtitleAPI    string = BilibiliAPI + "/m/subtitle"
	BilibiliTimelineAPI    string = BilibiliAPI + "/web/v2/home/timeline"
	BilibiliSearchAPI      string = BilibiliAPI + "/web/v2/search_v2/anime"
	// BilibiliSubtitleAPI string = bilibiliAPI + "/subtitle?s_locale&episode_id="
)

func GetApi[S Info | Episodes | Episode | EpisodeFile | Timeline | Search](s *S, url string, query map[string]string) (*S, error) {
	resp, err := utils.Request(url, query)
	if err != nil {
		return nil, err
	}

	if err = utils.JsonUnmarshal(resp, s); err != nil {
		return nil, err
	}

	return s, nil
}

func GetSubtitle(url, fileType string) ([]byte, error) {
	resp, err := utils.Request(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	switch fileType {
	case ".srt":
		subJson := new(Subtitle)
		if err := utils.JsonUnmarshal(resp, subJson); err != nil {
			return nil, err
		}

		return []byte(subJson.toSRT()), nil
	default:
		body, err := io.ReadAll(resp)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
}

func (subJson *Subtitle) toSRT() string {
	sub := make([]string, 0, len(subJson.Body))
	for i, s := range subJson.Body {
		content := s.Content
		if s.Location != 2 {
			content = fmt.Sprintf("{\\an%d}%s", s.Location, content)
		}
		sub = append(sub, fmt.Sprintf("%d\n%s --> %s\n%s", i+1, utils.SecondToTime(s.From), utils.SecondToTime(s.To), content))
	}
	return strings.Join(sub, "\n\n") + "\n"
}

// func ExtractSel[E Section | Episode](e []E, sel []string) []E {
// 	if sel == nil {
// 		return e
// 	}

// 	selIndex := utils.ListSelect(sel, len(e))
// 	sec := make([]E, 0, len(selIndex))
// 	for i, s := range e {
// 		if slices.Contains(selIndex, i+1) {
// 			sec = append(sec, s)
// 		}
// 	}
// 	return sec
// }

func ExtractEp(sections []Section, secSel []string, epSel []string) []Episode {
	var maxEp int
	secSelect := utils.ListSelect(secSel, len(sections))
	eps := []Episode{}
	for si, ss := range sections {
		if secSel == nil || slices.Contains(secSelect, si+1) {
			epSelect := utils.ListSelect(epSel, maxEp+len(ss.Episodes))
			for ei, es := range ss.Episodes {
				if epSel == nil || slices.Contains(epSelect, maxEp+ei+1) {
					eps = append(eps, es)
				}
			}
			maxEp += len(ss.Episodes)
		}
	}
	return eps
}
