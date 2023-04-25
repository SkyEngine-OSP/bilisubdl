package bilibili

import (
	"encoding/json"
	"time"
)

type Info struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Season struct {
			Title string `json:"title"`
		} `json:"season"`
	} `json:"data"`
}

type Episodes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Sections []Section `json:"sections"`
	} `json:"data"`
}

type Section struct {
	Title       string    `json:"title"`
	EpListTitle string    `json:"ep_list_title"`
	Episodes    []Episode `json:"episodes"`
}

type Episode struct {
	ShortTitleDisplay string      `json:"short_title_display"`
	LongTitleDisplay  string      `json:"long_title_display"`
	EpisodeID         json.Number `json:"episode_id"`
	TitleDisplay      string      `json:"title_display"`
	PublishTime       time.Time   `json:"publish_time"`
}

type EpisodeFile struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Subtitles []struct {
			URL string `json:"url"`
			ID  int    `json:"id"`
			// Lang    string `json:"lang"`
			Title string `json:"title"`
			// LangKey string `json:"lang_key"`
			Key       string `json:"key"`
			IsMachine bool   `json:"is_machine"`
		} `json:"subtitles"`
	} `json:"data"`
}

type Subtitle struct {
	Body []struct {
		From     float64 `json:"from"`
		To       float64 `json:"to"`
		Location int     `json:"location"`
		Content  string  `json:"content"`
	} `json:"body"`
}

type Timeline struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	// TTL     int    `json:"ttl"`
	Data struct {
		Items []struct {
			DayOfWeek string `json:"day_of_week"`
			IsToday   bool   `json:"is_today"`
			// DateText     string `json:"date_text"`
			FullDateText string `json:"full_date_text"`
			Cards        []struct {
				// Type        string      `json:"type"`
				// CardType    string      `json:"card_type"`
				Title string `json:"title"`
				// Cover       string      `json:"cover"`
				// View        string      `json:"view"`
				// Styles      string      `json:"styles"`
				// StyleList   interface{} `json:"style_list"`
				SeasonID string `json:"season_id"`
				// EpisodeID   string      `json:"episode_id"`
				IndexShow string `json:"index_show"`
				// Label       int         `json:"label"`
				// RankInfo    interface{} `json:"rank_info"`
				// ViewHistory interface{} `json:"view_history"`
				// Watched     string      `json:"watched"`
				// Duration    string      `json:"duration"`
				// ViewAt      string      `json:"view_at"`
				// PubTimeText string `json:"pub_time_text"`
				// Unavailable bool        `json:"unavailable"`
			} `json:"cards"`
		} `json:"items"`
	} `json:"data"`
}

type Search struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	// TTL     int    `json:"ttl"`
	Data struct {
		Items []struct {
			Title    string `json:"title"`
			SeasonID string `json:"season_id"`
			// Highlights []struct {
			// 	Str   string `json:"str"`
			// 	Match bool   `json:"match"`
			// } `json:"highlights"`
			// Cover          string `json:"cover"`
			// View           string `json:"view"`
			// SeasonType     string `json:"season_type"`
			// SeasonTypeEnum int    `json:"season_type_enum"`
			// Styles         []struct {
			// 	ID    int    `json:"id"`
			// 	Title string `json:"title"`
			// 	Qs    string `json:"qs"`
			// } `json:"styles"`
			// Description   string `json:"description"`
			// PayPolicyEnum int    `json:"pay_policy_enum"`
			// ContentRating int    `json:"content_rating"`
			// UpdatePattern string `json:"update_pattern"`
			// Label         int    `json:"label"`
			IndexShow string `json:"index_show"`
		} `json:"items"`
		// Qid     string `json:"qid"`
		// HasNext bool   `json:"has_next"`
	} `json:"data"`
}
