package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/K0ng2/bilisubdl/pkg/bilibili"
	"github.com/K0ng2/bilisubdl/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	language      string
	output        string
	listLang      bool
	listSection   bool
	overwrite     bool
	timeline      string
	search        string
	sectionSelect []string
	episodeSelect []string
)

var RootCmd = &cobra.Command{
	Use: "bilisubdl [id] [flags]",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch {
		case listLang:
			err = RunListLanguage(args[0])
		case listSection:
			err = RunListSection(args[0])
		case timeline != "":
			err = RunTimeline()
		case search != "":
			err = RunSearch()
		default:
			for _, s := range args {
				err := Run(s)
				if err != nil {
					fmt.Fprintln(os.Stderr, "ID:", s, err)
				}
			}
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
	Example: "bilisubdl 37738 1042594 -l th\nbilisubdl 37738 --list-subs\nbilisubdl --timeline=sun",
}

func init() {
	rootFlags := RootCmd.PersistentFlags()
	rootFlags.StringVarP(&language, "language", "l", "", "Subtitle language to download (e.g. en)")
	rootFlags.StringVarP(&output, "output", "o", "./", "Set output")
	rootFlags.BoolVarP(&listLang, "list-language", "L", false, "List available subtitle language")
	rootFlags.BoolVar(&listSection, "list-section", false, "List available section")
	rootFlags.BoolVarP(&overwrite, "overwrite", "w", false, "Force overwrite downloaded subtitles")
	rootFlags.StringVarP(&search, "search", "s", "", "Search anime")
	rootFlags.StringVarP(&timeline, "timeline", "T", "", "Show timeline (sun|mon|tue|wed|thu|fri|sat)")
	rootFlags.Lookup("timeline").NoOptDefVal = "today"
	rootFlags.StringArrayVar(&sectionSelect, "section", nil, "Section select (e.g. 5,8-10)")
	rootFlags.StringArrayVar(&episodeSelect, "episode", nil, "Episode select (e.g. 5,8-10)")
}

func Run(id string) error {
	var (
		title, filename, fileType string
		episode                   *bilibili.Episode
		sub                       []byte
		exist                     bool
		maxEp                     int
	)
	info, err := bilibili.GetInfo(id)
	if err != nil {
		return err
	}

	epList, err := bilibili.GetEpisodes(id)
	if err != nil {
		return err
	}

	title = utils.CleanText(info.Data.Season.Title)
	if !listLang {
		err = os.MkdirAll(filepath.Join(output, title), os.ModePerm)
		if err != nil {
			return err
		}
	}

	sectionIndex := utils.ListSelect(sectionSelect, len(epList.Data.Sections))
	for ji, j := range epList.Data.Sections {
		if sectionSelect != nil && !slices.Contains(sectionIndex, ji+1) {
			continue
		}
		episodeIndex := utils.ListSelect(episodeSelect, maxEp+len(j.Episodes))
		for si, s := range j.Episodes {
			if episodeSelect != nil && !slices.Contains(episodeIndex, maxEp+si+1) {
				continue
			}
			filename = filepath.Join(output, title, fmt.Sprintf("%s.%s", utils.CleanText(s.TitleDisplay), language))
			for _, k := range []string{".srt", ".ass"} {
				if _, err := os.Stat(filename + k); err == nil && !overwrite {
					fmt.Println("#", filename+k)
					exist = true
					continue
				}
			}

			if exist {
				exist = false
				continue
			}

			episode, err = bilibili.GetEpisode(s.EpisodeID.String())
			if err != nil {
				fmt.Println(err)
			}

			sub, fileType, err = episode.Subtitle(language)
			if err != nil {
				return err
			}

			err := utils.WriteFile(filename+fileType, sub, s.PublishTime)
			if err != nil {
				return err
			}
			fmt.Println("*", filename+fileType)
		}
		maxEp += len(j.Episodes)
	}
	return nil
}

func RunTimeline() error {
	tl, err := bilibili.GetTimeline()
	if err != nil {
		return err
	}
	table := newTable(nil)
	for _, s := range tl.Data.Items {
		if timeline == "today" && s.IsToday {
			timeline = s.DayOfWeek
		}
		if s.DayOfWeek == strings.ToUpper(timeline) {
			for _, j := range s.Cards {
				table.Append([]string{j.SeasonID, j.Title, j.IndexShow})
			}
			table.SetHeader([]string{"ID", fmt.Sprintf("Title (%s %s)", s.DayOfWeek, s.FullDateText), "Status"})
			break
		}
	}
	table.Render()
	return nil
}

func RunSearch() error {
	ss, err := bilibili.GetSearch(search, "10")
	if err != nil {
		return err
	}
	table := newTable([]string{"ID", "Title", "Status"})
	for _, j := range ss.Data {
		if j.Module == "ogv" || j.Module == "ogv_subject" {
			for _, s := range j.Items {
				table.Append([]string{s.SeasonID.String(), s.Title, s.IndexShow})
			}
			break
		}
	}
	table.Render()
	return nil
}

func RunListLanguage(id string) error {
	info, err := bilibili.GetInfo(id)
	if err != nil {
		return err
	}

	epList, err := bilibili.GetEpisodes(id)
	if err != nil {
		return err
	}

	episode, err := bilibili.GetEpisode(epList.Data.Sections[0].Episodes[0].EpisodeID.String())
	if err != nil {
		return err
	}

	table := newTable([]string{"Key", "Lang"})
	for _, s := range episode.Data.Subtitles {
		table.Append([]string{s.Key, s.Title})
	}
	fmt.Println("Title:", info.Data.Season.Title)
	table.Render()
	return nil
}

func RunListSection(id string) error {
	info, err := bilibili.GetInfo(id)
	if err != nil {
		return err
	}
	epList, err := bilibili.GetEpisodes(id)
	if err != nil {
		return err
	}
	table := newTable([]string{"index", "section"})
	for i, s := range epList.Data.Sections {
		table.Append([]string{strconv.Itoa(i + 1), s.EpListTitle})
	}
	fmt.Println("Title:", info.Data.Season.Title)
	table.Render()
	return nil
}

func newTable(header []string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetHeader(header)
	return table
}
