package implementations

import (
	"fmt"
	"strings"

	models "github.com/alejandro-bustamante/flick/internal/models"
)

type Extractor struct {
	yearRange       [2]int
	qualityKeywords []string
	sourceKeywords  []string
	langKeywords    []string
}

func NewExtractor() *Extractor {
	return &Extractor{
		yearRange:       [2]int{1900, 2030},
		qualityKeywords: []string{"720p", "1080p", "4K", "2160p", "480p", "BluRay", "BRRip"},
		sourceKeywords:  []string{"BluRay", "WEB-DL", "HDRip", "HDTV", "Netflix", "AMZN"},
		langKeywords:    []string{"SPANISH", "LATINO", "ENGLISH", "DUAL"},
	}
}

func (e *Extractor) Extract(tokens []string) *models.MediaInfo {
	info := &models.MediaInfo{}
	var titleTokens []string
	var trailing []string

	lastRelevantIdx := -1

	// Buscar año o temporada/episodio
	for i, token := range tokens {
		if year := e.extractYear(token); year > 0 {
			info.Year = year
			info.IsMovie = true
			lastRelevantIdx = i
		}

		if season, episode := e.extractSeasonEpisode(token); season > 0 {
			info.Season = season
			info.Episode = episode
			info.IsMovie = false
			lastRelevantIdx = i
		}
	}

	// Determinar título y "basura útil"
	if lastRelevantIdx >= 0 {
		titleTokens = tokens[:lastRelevantIdx]
		trailing = tokens[lastRelevantIdx+1:]
	} else {
		titleTokens = tokens
	}

	info.Title = e.buildTitle(titleTokens)
	info.Remaining = trailing // puedes usar esto para debugging o fallback

	return info
}

func (e *Extractor) extractYear(token string) int {
	if len(token) == 4 {
		var year int
		if _, err := fmt.Sscanf(token, "%d", &year); err == nil {
			if year >= e.yearRange[0] && year <= e.yearRange[1] {
				return year
			}
		}
	}
	return 0
}

func (e *Extractor) extractSeasonEpisode(token string) (int, int) {
	upper := strings.ToUpper(token)

	if strings.HasPrefix(upper, "S") && strings.Contains(upper, "E") {
		var season, episode int
		if _, err := fmt.Sscanf(upper, "S%dE%d", &season, &episode); err == nil {
			return season, episode
		}
	}

	return 0, 0
}

func (e *Extractor) buildTitle(tokens []string) string {
	return strings.Join(tokens, " ")
}
