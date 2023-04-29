package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/K0ng2/bilisubdl/pkg/bilibili"
	"github.com/K0ng2/bilisubdl/utils"
	"github.com/fatih/color"
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
	fastCheck     bool
	skipMachine   bool
	dlArchive     string
	epFilename    string
	sectionSelect []string
	episodeSelect []string
)

var RootCmd = &cobra.Command{
	Use: "bilisubdl",
}

var dlCmd = &cobra.Command{
	Use:     "dl [ID] [flags]",
	Short:   "command downloads the subtitle for the given anime ID.",
	Args:    cobra.MinimumNArgs(1),
	Example: "bilisubdl dl 37738 1042594 -l th -o /path/to/output",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dlepisode {
			return runDlEpisode(args)
		} else {
			for _, s := range args {
				err := runDl(s)
				if err != nil {
					return fmt.Errorf("[ID: %s] %w", s, err)
				}
			}
		}

		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [keyword] [flags]",
	Short: "command allows you to search for anime on Bilibili based on a keyword.",
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
	Example: "bilisubdl search \"Attack on Titan\"\nbilisubdl search \"One Piece\" --json",
}

var timelineCmd = &cobra.Command{
	Use:   "timeline [day] [flags]",
	Short: "command allows you to view a timeline of Bilibili videos uploaded on a specific day of the week.",
	Long: `
The day of the week for which you want to view the timeline.
Can be one of 'sun', 'mon', 'tue', 'wed', 'thu', 'fri', or 'sat'.
If no day is provided, the current day of the week is used.
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runTimeline("")
		}

		return runTimeline(args[0])
	},
	Example: "bilisubdl timeline\nbilisubdl timeline wed --json",
}

var listCmd = &cobra.Command{
	Use:   "list [ID] [flags]",
	Short: "command allows you to view information about available subtitles and episodes for a given anime ID",
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
		return runList(args[0])
	},
}

func init() {
	RootCmd.AddCommand(dlCmd, searchCmd, timelineCmd, listCmd)
	selectFlags := flag.NewFlagSet("selectFlags", flag.ExitOnError)
	selectFlags.StringArrayVar(&sectionSelect, "section-range", nil, "selects a range of episodes to download subtitles for (e.g., `5`, `8-10`).")
	selectFlags.StringArrayVar(&episodeSelect, "episode-range", nil, "selects a range of sections to download subtitles for (e.g., `5`, `8-10`).")

	dlFlag := dlCmd.PersistentFlags()
	dlFlag.StringVarP(&language, "language", "l", "", "sets the subtitle language to download (e.g., `en` for English, `zh` for Chinese).")
	dlFlag.StringVarP(&output, "output", "o", "./", "sets the output directory where the downloaded subtitle file will be saved (default is the current directory).")
	dlFlag.BoolVar(&dlepisode, "dlepisode", false, "downloads the subtitle for the specified episode ID.")
	dlFlag.StringVar(&epFilename, "filename", "", "sets the subtitle filename using a specified format. This option only works in combination with `--dlepisode` flag. (e.g. Abc %d = Abc 1, Abc %02d = Abc 02)")
	dlFlag.BoolVarP(&overwrite, "overwrite", "w", false, "forces the tool to overwrite existing subtitle files in the output directory.")
	dlFlag.BoolVarP(&quiet, "quiet", "q", false, "suppresses verbose output.")
	dlFlag.BoolVar(&skipMachine, "skip-machine", false, "skips Machine translation.")
	dlFlag.AddFlagSet(selectFlags)
	dlFlag.BoolVar(&fastCheck, "fast-check", false, "skips checking the subtitle extension from API.")
	dlFlag.StringVar(&dlArchive, "download-archive", "", "Create a FILE to keep track of all downloaded and skipped subtitles, and use it to prevent downloading any files that are already recorded in it. Additionally, record the IDs of all newly downloaded subtitles in the same FILE.")
	dlCmd.MarkPersistentFlagRequired("language")
	dlCmd.MarkFlagsRequiredTogether("filename", "dlepisode")
	dlCmd.MarkFlagsMutuallyExclusive("fast-check", "overwrite")

	shareFlags := flag.NewFlagSet("shareFlags", flag.ExitOnError)
	shareFlags.BoolVar(&isJson, "json", false, "displays the output in JSON format.")
	searchFlag := searchCmd.PersistentFlags()
	searchFlag.AddFlagSet(shareFlags)

	timelineFlag := timelineCmd.PersistentFlags()
	timelineFlag.AddFlagSet(shareFlags)

	listFlag := listCmd.PersistentFlags()
	listFlag.BoolVarP(&listLang, "language", "L", false, "lists available subtitle languages for the specified anime ID.")
	listFlag.BoolVarP(&listSection, "section", "S", false, "lists available sections for the specified anime ID.")
	listFlag.BoolVarP(&listEpisode, "episode", "E", false, "lists available episodes for the specified anime ID.")
	listFlag.AddFlagSet(selectFlags)
	listCmd.MarkFlagsMutuallyExclusive("language", "section", "episode")
	listCmd.MarkFlagsMutuallyExclusive("language", "section-range", "episode-range")
}

func runDl(id string) error {
	var (
		title, filename string
		maxEp           int
	)

	query := map[string]string{
		"s_locale":  "en_US",
		"season_id": id,
	}

	info, err := bilibili.GetApi(new(bilibili.Info), bilibili.BilibiliSeasonInfoAPI, query)
	if err != nil {
		return err
	}

	epList, err := bilibili.GetApi(new(bilibili.Episodes), bilibili.BilibiliEpisodeInfoAPI, query)
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
					filename = filepath.Join(title, fmt.Sprintf("%s.%s", utils.CleanText(s.TitleDisplay), language))

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

		if err := downloadSub(id, filename, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

func downloadSub(episodeId, filename string, publishTime time.Time) error {
	outFile := filepath.Join(output, filename)

	if fastCheck {
		for _, k := range []string{".srt", ".ass"} {
			if _, err := os.Stat(outFile + k); !os.IsNotExist(err) && !overwrite && !quiet {
				fmt.Println(color.HiBlackString("# %s", filename+k), color.HiYellowString("fast-check"))
				return nil
			}
		}
	}

	episode, err := bilibili.GetApi(new(bilibili.EpisodeFile), bilibili.BilibiliSubtitleAPI, map[string]string{"ep_id": episodeId})
	if err != nil {
		return err
	}

	for _, k := range episode.Data.Subtitles {
		if k.Key == language {
			if k.IsMachine {
				if skipMachine {
					color.Yellow("- %s", filename)
					return nil
				}
				color.Red("Warning: The downloaded subtitle has been machine translated and may contain errors or inaccuracies")
			}

			fileType := filepath.Ext(strings.Split(k.URL, "?")[0])
			if fileType == ".json" {
				fileType = ".srt"
			}
			if dlArchive != "" {
				isInArchive, err := checkArchive(strconv.Itoa(k.ID))
				if err != nil {
					return err
				}
				if isInArchive && !overwrite {
					fmt.Println(color.HiBlackString("# %s", filename+fileType), color.HiYellowString("archive"))
					return nil
				}
				if _, err := os.Stat(outFile + fileType); !os.IsNotExist(err) && !overwrite {
					err = add2Archive(strconv.Itoa(k.ID))
					if err != nil {
						return err
					}
					fmt.Println(color.HiBlackString("# %s", filename+fileType), color.HiYellowString("exist, add to archive"))
					return nil
				}
			} else if _, err := os.Stat(outFile + fileType); !os.IsNotExist(err) && !overwrite {
				fmt.Println(color.HiBlackString("# %s", filename+fileType), color.HiYellowString("exist"))
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(outFile), 0o700); err != nil {
				return err
			}

			sub, err := bilibili.GetSubtitle(k.URL, fileType)
			if err != nil {
				return err
			}

			if err := utils.WriteFile(outFile+fileType, sub, publishTime); err != nil {
				return err
			}

			if dlArchive != "" {
				isInArchive, err := checkArchive(strconv.Itoa(k.ID))
				if err != nil {
					return err
				}
				if !isInArchive {
					err = add2Archive(strconv.Itoa(k.ID))
					if err != nil {
						return err
					}
				}
			}

			if !quiet {
				color.Green("* %s", filename+fileType)
			}
		}
		continue
	}
	return nil
}

func runTimeline(day string) error {
	tl, err := bilibili.GetApi(new(bilibili.Timeline), bilibili.BilibiliTimelineAPI, nil)
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
	query := map[string]string{
		"keyword":  s,
		"platform": "web",
		"pn":       "1",
		"ps":       "20",
		"s_locale": "en_US",
	}

	ss, err := bilibili.GetApi(new(bilibili.Search), bilibili.BilibiliSearchAPI, query)
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
		for _, j := range ss.Data.Items {
			table.Append([]string{j.SeasonID, j.Title, j.IndexShow})
		}
		if table.NumLines() == 0 {
			fmt.Println("No results found for your search query. Please try a different keyword or check your spelling.")
		} else {
			table.Render()
		}
	}
	return nil
}

func runList(id string) error {
	query := map[string]string{
		"s_locale":  "en_US",
		"season_id": id,
	}

	info, err := bilibili.GetApi(new(bilibili.Info), bilibili.BilibiliSeasonInfoAPI, query)
	if err != nil {
		return err
	}

	epList, err := bilibili.GetApi(new(bilibili.Episodes), bilibili.BilibiliEpisodeInfoAPI, query)
	if err != nil {
		return err
	}

	if len(epList.Data.Sections) == 0 {
		return fmt.Errorf("The list is currently empty. Please check back later.")
	}

	fmt.Println("Title:", info.Data.Season.Title)
	table := newTable(nil)

	switch {
	case listLang:
		episode, err := bilibili.GetApi(new(bilibili.EpisodeFile), bilibili.BilibiliSubtitleAPI, map[string]string{"ep_id": epList.Data.Sections[0].Episodes[0].EpisodeID.String()})

		if err != nil {
			return err
		}

		table.SetHeader([]string{"Key", "Lang"})
		for _, s := range episode.Data.Subtitles {
			table.Append([]string{s.Key, s.Title})
		}
	case listSection:
		table.SetHeader([]string{"#", "episode", "title"})
		for i, s := range epList.Data.Sections {
			table.Append([]string{strconv.Itoa(i + 1), s.EpListTitle, s.Title})
		}
	case listEpisode:
		table.SetHeader([]string{"#", "title"})
		for _, s := range bilibili.ExtractEp(epList.Data.Sections, sectionSelect, episodeSelect) {
			table.Append([]string{s.ShortTitleDisplay, s.LongTitleDisplay})
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

func checkArchive(subtitleID string) (bool, error) {
	f, err := os.Open(dlArchive)
	defer f.Close()
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if subtitleID == scanner.Text() {
			return true, nil
		}
	}
	return false, nil
}

func add2Archive(subtitleID string) error {
	f, err := os.OpenFile(dlArchive, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o700)
	defer f.Close()
	if err != nil {
		return err
	}

	if _, err := f.WriteString(subtitleID + "\n"); err != nil {
		return err
	}

	return nil
}
