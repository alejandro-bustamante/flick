package parser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	models "github.com/alejandro-bustamante/flick/internal/models"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Logger para debugging y testing (importado desde utils)
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type MediaParser struct {
	logger         Logger
	separators     []string
	junkPatterns   map[string]struct{}
	yearRange      [2]int
	minTitleLength int
}

// NewMediaParser ahora recibe la configuración directamente
func NewMediaParser(separators []string, junkPatterns []string, yearRange []int, l Logger) *MediaParser {

	// Procesar junkPatterns (lógica que estaba en NewCleaner)
	junk := make(map[string]struct{}, len(junkPatterns))
	for _, p := range junkPatterns {
		junk[strings.ToLower(p)] = struct{}{}
	}

	return &MediaParser{
		logger:         l,
		separators:     separators,
		junkPatterns:   junk,
		yearRange:      [2]int(yearRange), // Lógica de NewExtractor
		minTitleLength: 1,                 // Lógica de NewValidator
	}
}

// The two Parse functions asume the <filename> recieved HAS NO EXTENSION
func (p *MediaParser) Parse(filename string) *models.ParseResult {
	result := &models.ParseResult{}

	p.logger.Debug("Parsing file: %s", filename)

	// Fase 1: Tokenización (interna)
	tokens := p.tokenize(filename)
	p.logger.Debug("Tokens: %v", tokens)

	// Fase 2: Limpieza (interna)
	cleanTokens := p.clean(tokens)
	p.logger.Debug("Clean tokens: %v", cleanTokens)

	// Fase 3: Extracción (interna)
	info := p.extract(cleanTokens)
	info.OriginalName = filename

	result.MediaInfo = info

	return result
}

func (p *MediaParser) ParseNormalized(filename string) *models.ParseResult {
	result := &models.ParseResult{}

	p.logger.Debug("Parsing file: %s", filename)

	ext := filepath.Ext(filename)
	filename = strings.Trim(filename, ext)

	// Stage 1: Tokenization (interna)
	tokens := p.tokenize(filename)
	p.logger.Debug("Tokens: %v", tokens)

	// Stage 2: Normalization (interna)
	normalizedTokens := p.normalizeTokens(tokens)
	p.logger.Debug("Normalized tokens: %v", normalizedTokens)

	// Stage 3: Cleaning (interna)
	cleanTokens := p.clean(normalizedTokens)
	p.logger.Debug("Clean tokens: %v", cleanTokens)

	// Stage 4: Extraction (interna)
	info := p.extract(cleanTokens)
	info.OriginalName = filename

	result.MediaInfo = info
	return result
}

func (p *MediaParser) tokenize(input string) []string {
	for _, sep := range p.separators {
		input = strings.ReplaceAll(input, sep, " ")
	}
	tokens := strings.Fields(input)
	return tokens
}

// --- Lógica de Cleaner ---
func (p *MediaParser) isJunk(token string) bool {
	_, exists := p.junkPatterns[strings.ToLower(token)]
	return exists
}

func (p *MediaParser) clean(tokens []string) []string {
	var cleaned []string
	for _, token := range tokens {
		if !p.isJunk(token) {
			cleaned = append(cleaned, token)
		}
	}
	return cleaned
}

// --- Lógica de Extractor ---
func (p *MediaParser) extract(tokens []string) *models.MediaInfo {
	info := &models.MediaInfo{}
	var titleTokens []string
	var trailing []string

	lastRelevantIdx := -1

	// Buscar año o temporada/episodio
	for i, token := range tokens {
		if year := p.extractYear(token); year > 0 {
			info.Year = year
			lastRelevantIdx = i
		}

		if season, episode := p.extractSeasonEpisode(token); season > 0 {
			info.Season = season
			info.Episode = episode
			lastRelevantIdx = i
			info.IsSeries = true
		}
	}

	// Dividir título del resto
	if lastRelevantIdx >= 0 {
		titleTokens = tokens[:lastRelevantIdx]
		trailing = tokens[lastRelevantIdx+1:]
	} else {
		titleTokens = tokens
	}

	info.Title = p.buildTitle(titleTokens)
	info.Remaining = trailing

	return info
}

func (p *MediaParser) extractYear(token string) int {
	if len(token) == 4 {
		if year, err := strconv.Atoi(token); err == nil {
			if year >= p.yearRange[0] && year <= p.yearRange[1] {
				return year
			}
		}
	}
	return 0
}

func (p *MediaParser) extractSeasonEpisode(token string) (int, int) {
	lower := strings.ToLower(token)

	seasonEpisodePatterns := []*regexp.Regexp{
		regexp.MustCompile(`^s(\d+)e(\d+)`),
		regexp.MustCompile(`^season(\d+)episode(\d+)`),
		regexp.MustCompile(`^s(\d+)ep(\d+)`),
		regexp.MustCompile(`^(\d+)x(\d+)`),
	}

	for _, pattern := range seasonEpisodePatterns {
		matches := pattern.FindStringSubmatch(lower)
		if len(matches) == 3 {
			season, err1 := strconv.Atoi(matches[1])
			episode, err2 := strconv.Atoi(matches[2])
			if err1 == nil && err2 == nil && season > 0 && episode > 0 {
				return season, episode
			}
		}
	}

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

func (p *MediaParser) buildTitle(tokens []string) string {
	return strings.Join(tokens, " ")
}

func (p *MediaParser) normalizeTokens(tokens []string) []string {
	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)

	for i, token := range tokens {
		text, _, _ := transform.String(t, token)
		tokens[i] = strings.ToLower(text)
	}
	return tokens
}

func (p *MediaParser) validate(info *models.MediaInfo) error {
	if err := p.validateTitle(info.Title); err != nil {
		return err
	}

	if info.Season == 0 {
		return fmt.Errorf("TV show must have season number")
	}

	return nil
}

func (p *MediaParser) validateTitle(title string) error {
	if len(strings.TrimSpace(title)) < p.minTitleLength {
		return fmt.Errorf("title too short: '%s'", title)
	}
	return nil
}

func (p *MediaParser) NormalizeForComparison(input string) string {
	// 1. Replace separators
	temp := input
	for _, sep := range p.separators {
		temp = strings.ReplaceAll(temp, sep, " ")
	}

	// 2. Takeout diacritics
	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
	normalized, _, _ := transform.String(t, temp)

	// 3. Collapse to spaces and lowercase
	processed := strings.Join(strings.Fields(normalized), " ")
	return strings.ToLower(processed)
}
