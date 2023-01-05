package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/K0ng2/bilisubdl/pkg/bilibili"
	"github.com/K0ng2/bilisubdl/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

var (
	language      string
	output        string
	listLang      bool
	listSection   bool
	listEpisode   bool
	overwrite     bool
	dlepisode     bool
	isJson        bool
	quiet         bool
	timeline      string
	search        string
	epFilename    string
	sectionSelect []string
	episodeSelect []string
)

var RootCmd = &cobra.Command{
	Use: "bilisubdl",
}

var dlCmd = &cobra.Command{
	Use:     "dl [ID] [flags]",
	Short:   "Download subtitle from ID.",
	Args:    cobra.MinimumNArgs(1),
	Example: "bilisubdl dl 37738 1042594 -l th",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, s := range args {
			err := runDl(s)
			if err != nil {
				return fmt.Errorf("[ID: %s] %w", s, err)
			}
		}
		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search anime",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runSearch(args[0])
		if err != nil {
			return fmt.Errorf("[keyword: %s] %w", args[0], err)
		}
		return nil
	},
}

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Show timeline (sun|mon|tue|wed|thu|fri|sat)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runTimeline("")
		}
		return runTimeline(args[0])
	},
	Example: "bilisubdl timeline\nbilisubdl timeline sun",
}

var listCmd = &cobra.Command{
	Use:   "list [ID]",
	Short: "Show info",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runList(args[0])
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(dlCmd, searchCmd, timelineCmd, listCmd)

	selectFlags := flag.NewFlagSet("selectFlags", flag.ExitOnError)
	selectFlags.StringArrayVar(&sectionSelect, "section-range", nil, "Section select (e.g. 5,8-10)")
	selectFlags.StringArrayVar(&episodeSelect, "episode-range", nil, "Episode select (e.g. 5,8-10)")

	dlFlag := dlCmd.PersistentFlags()
	dlFlag.StringVarP(&language, "language", "l", "", "Subtitle language to download (e.g. en)")
	dlFlag.StringVarP(&output, "output", "o", "./", "Set output directory")
	dlFlag.BoolVar(&dlepisode, "dlepisode", false, "Download subtitle from episode id")
	dlFlag.StringVar(&epFilename, "filename", "", "Set subtitle filename (e.g. Abc %d = Abc 1, Abc %02d = Abc 02)\n(This option only works in combination with --dlepisode flag)")
	dlFlag.BoolVarP(&overwrite, "overwrite", "w", false, "Force overwrite downloaded subtitles")
	dlFlag.BoolVarP(&quiet, "quiet", "q", false, "Quiet verbose")
	dlFlag.AddFlagSet(selectFlags)
	dlCmd.MarkFlagRequired("language")
	dlCmd.MarkFlagsRequiredTogether("filename", "dlepisode")

	shareFlags := flag.NewFlagSet("shareFlags", flag.ExitOnError)
	shareFlags.BoolVar(&isJson, "json", false, "Display in JSON format.")
	searchFlag := searchCmd.PersistentFlags()
	searchFlag.AddFlagSet(shareFlags)

	timelineFlag := timelineCmd.PersistentFlags()
	timelineFlag.AddFlagSet(shareFlags)

	listFlag := listCmd.PersistentFlags()
	listFlag.BoolVarP(&listLang, "language", "L", false, "List available subtitle language")
	listFlag.BoolVarP(&listSection, "section", "S", false, "List available section")
	listFlag.BoolVarP(&listEpisode, "episode", "E", false, "List available episode")
	listFlag.AddFlagSet(selectFlags)
	listCmd.MarkFlagsMutuallyExclusive("language", "section", "episode")
	listCmd.MarkFlagsMutuallyExclusive("language", "section-range", "episode-range")
}

func runDl(id string) error {
	var (
		title, filename string
		maxEp           int
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
	sectionIndex := utils.ListSelect(sectionSelect, len(epList.Data.Sections))
	for ji, j := range epList.Data.Sections {
		if sectionSelect == nil || slices.Contains(sectionIndex, ji+1) {
			episodeIndex := utils.ListSelect(episodeSelect, maxEp+len(j.Episodes))
			for si, s := range j.Episodes {
				if episodeSelect == nil || slices.Contains(episodeIndex, maxEp+si+1) {
					filename = filepath.Join(output, title, fmt.Sprintf("%s.%s", utils.CleanText(s.TitleDisplay), language))

					if err := downloadSub(s.EpisodeID.String(), filename, s.PublishTime); err != nil {
						return err
					}
				}
			}
			maxEp += len(j.Episodes)
		}
	}
	return nil
}

func runDlEpisode(ids []string) error {
	var filename string
	if output != "" {
		if err := os.MkdirAll(output, 0700); os.IsExist(err) {
			return err
		}
	}

	for i, id := range ids {
		filename = id
		if epFilename != "" {
			filename = fmt.Sprintf(epFilename, i+1)
		}
		filename = filepath.Join(output, filename)

		if err := downloadSub(id, filename, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

func downloadSub(id, filename string, publishTime time.Time) error {
	for _, k := range []string{".srt", ".ass"} {
		if _, err := os.Stat(filename + k); !os.IsNotExist(err) && !overwrite {
			if !quiet {
				fmt.Println("#", filename+k)
			}
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Join(filepath.Dir(filename)), 0700); os.IsExist(err) {
		return err
	}

	episode, err := bilibili.GetEpisode(id)
	if episode == nil {
		return err
	}

	sub, fileType, err := episode.Subtitle(language)
	if err != nil {
		return err
	}

	if err := utils.WriteFile(filename+fileType, sub, publishTime); err != nil {
		return err
	}

	if !quiet {
		fmt.Println("*", filename+fileType)
	}
	return nil
}

func runTimeline(day string) error {
	tl, err := bilibili.GetTimeline()
	if err != nil {
		return err
	}
	if isJson {
		b, err := json.Marshal(tl)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	} else {
		for _, s := range tl.Data.Items {
			if day == "" && s.IsToday {
				day = s.DayOfWeek
			}
			if s.DayOfWeek == strings.ToUpper(day) {
				if len(s.Cards) == 0 {
					fmt.Println("No updates")
				} else {
					table := newTable(nil)
					for _, j := range s.Cards {
						table.Append([]string{j.SeasonID, j.Title, j.IndexShow})
					}
					table.SetHeader([]string{"ID", fmt.Sprintf("Title (%s %s)", s.DayOfWeek, s.FullDateText), "Status"})
					table.Render()
					break
				}
			}
		}
	}
	return nil
}

func runSearch(s string) error {
	ss, err := bilibili.GetSearch(s, "10")
	if err != nil {
		return err
	}
	if isJson {
		b, err := json.Marshal(ss)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	} else {
		table := newTable([]string{"ID", "Title", "Status"})
		for _, j := range ss.Data {
			if j.Module == "ogv" || j.Module == "ogv_subject" {
				for _, s := range j.Items {
					table.Append([]string{s.SeasonID.String(), s.Title, s.IndexShow})
				}
				break
			}
		}
		if table.NumLines() == 0 {
			fmt.Println("No relevant results were found.")
		} else {
			table.Render()
		}
	}
	return nil
}

func runList(ID string) error {
	info, err := bilibili.GetInfo(ID)
	if err != nil {
		return err
	}

	epList, err := bilibili.GetEpisodes(ID)
	if err != nil {
		return err
	}

	if len(epList.Data.Sections) == 0 {
		return fmt.Errorf("Episode list not found Or not yet aired")
	}

	fmt.Println("Title:", info.Data.Season.Title)
	table := newTable(nil)

	switch {
	case listLang:
		episode, err := bilibili.GetEpisode(epList.Data.Sections[0].Episodes[0].EpisodeID.String())
		if err != nil {
			return err
		}
		table.SetHeader([]string{"Key", "Lang"})
		for _, s := range episode.Data.Subtitles {
			table.Append([]string{s.Key, s.Title})
		}
	case listSection:
		table.SetHeader([]string{"#", "section", "title"})
		sectionIndex := utils.ListSelect(sectionSelect, len(epList.Data.Sections))
		for i, s := range epList.Data.Sections {
			if sectionSelect == nil || slices.Contains(sectionIndex, i+1) {
				table.Append([]string{strconv.Itoa(i + 1), s.EpListTitle, s.Title})
			}
		}
	case listEpisode:
		var maxEp int
		table.SetHeader([]string{"#", "episode", "title"})
		sectionIndex := utils.ListSelect(sectionSelect, len(epList.Data.Sections))
		for ji, j := range epList.Data.Sections {
			if sectionSelect == nil || slices.Contains(sectionIndex, ji+1) {
				episodeIndex := utils.ListSelect(episodeSelect, maxEp+len(j.Episodes))
				for si, s := range j.Episodes {
					if episodeSelect == nil || slices.Contains(episodeIndex, maxEp+si+1) {
						table.Append([]string{s.ShortTitleDisplay, strconv.Itoa(ji + 1), s.LongTitleDisplay})
					}
				}
				maxEp += len(j.Episodes)
			}
		}
	}
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
