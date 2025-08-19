package transformations

import (
	"regexp"
	"strconv"
	"strings"

	models "github.com/alejandro-bustamante/flick/internal/models"
)

type Extractor struct {
	yearRange [2]int
}

func NewExtractor(yearRange []int) *Extractor {
	return &Extractor{
		yearRange: [2]int(yearRange),
	}
}

func (e *Extractor) Extract(tokens []string) *models.MediaInfo {
	info := &models.MediaInfo{}
	var titleTokens []string
	var trailing []string

	lastRelevantIdx := -1

	// Buscar año o temporada/episodio
	// Search for year (movie) or season/episode (series)
	for i, token := range tokens {
		if year := e.extractYear(token); year > 0 {
			info.Year = year
			lastRelevantIdx = i
		}

		if season, episode := e.extractSeasonEpisode(token); season > 0 {
			info.Season = season
			info.Episode = episode
			lastRelevantIdx = i
			info.IsSeries = true
		}
	}

	// Divide title from trailing string (probably junk)
	// Title(YEAR/SEASON-EPISODE).trailing string
	//         |--->lastRelevantIndex
	if lastRelevantIdx >= 0 {
		titleTokens = tokens[:lastRelevantIdx]
		trailing = tokens[lastRelevantIdx+1:]
	} else {
		titleTokens = tokens
	}

	info.Title = e.buildTitle(titleTokens)
	info.Remaining = trailing

	return info
}

func (e *Extractor) extractYear(token string) int {
	if len(token) == 4 {
		// if _, err := fmt.Sscanf(token, "%d", &year); err == nil {
		if year, err := strconv.Atoi(token); err == nil {
			if year >= e.yearRange[0] && year <= e.yearRange[1] {
				return year
			}
		}
	}
	return 0
}

func (e *Extractor) extractSeasonEpisode(token string) (int, int) {
	lower := strings.ToLower(token)

	seasonEpisodePatterns := []*regexp.Regexp{
		regexp.MustCompile(`^s(\d+)e(\d+)`),
		regexp.MustCompile(`^season(\d+)episode(\d+)`),
		regexp.MustCompile(`^s(\d+)ep(\d+)`),
		regexp.MustCompile(`^(\d+)x(\d+)`),
	}

	for _, pattern := range seasonEpisodePatterns {
		matches := pattern.FindStringSubmatch(lower)
		// matches[0] = full pattern found
		// matches[n] = n group inside the regex

		//Verify if found full pattern + season + episode
		if len(matches) == 3 {
			season, err1 := strconv.Atoi(matches[1])
			episode, err2 := strconv.Atoi(matches[2])
			if err1 == nil && err2 == nil && season > 0 && episode > 0 {
				return season, episode
			}
		}
	}

	// We need to search for some extra info to determine the season
	// Currently we asume season 1
	episodeOnlyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^e(\d+)`),
		regexp.MustCompile(`^episode(\d+)`),
		regexp.MustCompile(`^ep(\d+)`),
	}
	for _, pattern := range episodeOnlyPatterns {
		matches := pattern.FindStringSubmatch(lower)
		if len(matches) == 2 {
			episode, err := strconv.Atoi(matches[1])
			if err == nil && episode > 0 {
				return 1, episode
			}
		}
	}

	return 0, 0
}

func (e *Extractor) buildTitle(tokens []string) string {
	return strings.Join(tokens, " ")
}
