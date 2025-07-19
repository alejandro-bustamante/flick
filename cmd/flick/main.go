// Strategy
// Archivo → Normalización → Tokenización → Limpieza → Extracción → Resultado
package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type MediaInfo struct {
	Title    string
	Year     int
	Season   int
	Episode  int
	IsMovie  bool
	Quality  string
	Source   string
	Language string
}

// Palabras clave que indican calidad, fuente, etc.
var qualityKeywords = []string{
	"720p", "1080p", "4K", "2160p", "480p", "BluRay", "BRRip", "DVDRip",
	"WEBRip", "HDRip", "CAMRip", "TS", "TC", "HDTV", "WEB-DL", "x264",
	"x265", "HEVC", "H264", "H265", "DTS", "AC3", "AAC", "FLAC",
}

var sourceKeywords = []string{
	"BluRay", "BRRip", "DVDRip", "WEBRip", "HDRip", "HDTV", "WEB-DL", "Netflix",
	"Amazon", "Hulu", "Disney", "HBO", "Prime", "NF", "AMZN", "HULU",
}

var languageKeywords = []string{
	"SPANISH", "LATINO", "CASTELLANO", "ENGLISH", "DUBBED", "SUB", "SUBS",
	"SUBTITULADO", "DUAL", "MULTI",
}

// Separadores comunes en nombres de archivos
var separators = []string{".", "-", "_", " ", "[", "]", "(", ")"}

func ParseMediaFilename(filename string) *MediaInfo {
	// Remover extensión
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Normalizar separadores
	normalized := normalizeSeparators(name)

	// Tokenizar
	tokens := tokenize(normalized)

	// Limpiar tokens
	cleanTokens := cleanTokens(tokens)

	// Extraer información
	info := extractMediaInfo(cleanTokens)

	return info
}

func normalizeSeparators(input string) string {
	result := input
	for _, sep := range separators {
		if sep != " " {
			result = strings.ReplaceAll(result, sep, " ")
		}
	}
	return result
}

func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder

	for _, char := range input {
		if unicode.IsSpace(char) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func cleanTokens(tokens []string) []string {
	var cleaned []string

	for _, token := range tokens {
		// Convertir a minúsculas para comparación
		lowerToken := strings.ToLower(token)

		// Saltar tokens vacíos
		if len(token) == 0 {
			continue
		}

		// Saltar tokens que son solo números o caracteres especiales
		if isJunkToken(token) {
			continue
		}

		cleaned = append(cleaned, token)
	}

	return cleaned
}

func isJunkToken(token string) bool {
	// Tokens que son solo números pequeños (generalmente parte de codificación)
	if num, err := strconv.Atoi(token); err == nil && num < 10 {
		return true
	}

	// Tokens muy cortos que no son útiles
	if len(token) <= 2 && !isYear(token) && !isSeasonEpisode(token) {
		return true
	}

	return false
}

func extractMediaInfo(tokens []string) *MediaInfo {
	info := &MediaInfo{}
	var titleTokens []string

	for i, token := range tokens {
		// Detectar año
		if year := extractYear(token); year > 0 {
			info.Year = year
			// Todo antes del año es título
			titleTokens = tokens[:i]
			// Procesar tokens después del año
			info.processTechnicalTokens(tokens[i+1:])
			break
		}

		// Detectar temporada/episodio
		if season, episode := extractSeasonEpisode(token); season > 0 {
			info.Season = season
			info.Episode = episode
			info.IsMovie = false
			// Todo antes de S##E## es título
			titleTokens = tokens[:i]
			// Procesar tokens después
			info.processTechnicalTokens(tokens[i+1:])
			break
		}

		// Si llegamos aquí, es parte del título
		titleTokens = append(titleTokens, token)
	}

	// Si no encontramos año ni temporada, asumir que es película
	if info.Year == 0 && info.Season == 0 {
		info.IsMovie = true
		// Buscar información técnica al final
		cutoff := len(tokens)
		for i := len(tokens) - 1; i >= 0; i-- {
			if isTechnicalToken(tokens[i]) {
				cutoff = i
			} else {
				break
			}
		}
		titleTokens = tokens[:cutoff]
		info.processTechnicalTokens(tokens[cutoff:])
	}

	// Limpiar y construir título
	info.Title = buildCleanTitle(titleTokens)

	return info
}

func extractYear(token string) int {
	if len(token) == 4 {
		if year, err := strconv.Atoi(token); err == nil {
			if year >= 1900 && year <= 2030 {
				return year
			}
		}
	}
	return 0
}

func extractSeasonEpisode(token string) (int, int) {
	upper := strings.ToUpper(token)

	// Formato S##E## o S#E#
	if strings.HasPrefix(upper, "S") && strings.Contains(upper, "E") {
		parts := strings.Split(upper[1:], "E")
		if len(parts) == 2 {
			season, err1 := strconv.Atoi(parts[0])
			episode, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				return season, episode
			}
		}
	}

	return 0, 0
}

func isYear(token string) bool {
	return extractYear(token) > 0
}

func isSeasonEpisode(token string) bool {
	season, _ := extractSeasonEpisode(token)
	return season > 0
}

func isTechnicalToken(token string) bool {
	upper := strings.ToUpper(token)

	// Verificar palabras clave
	for _, keyword := range qualityKeywords {
		if strings.Contains(upper, strings.ToUpper(keyword)) {
			return true
		}
	}

	for _, keyword := range sourceKeywords {
		if strings.Contains(upper, strings.ToUpper(keyword)) {
			return true
		}
	}

	for _, keyword := range languageKeywords {
		if strings.Contains(upper, strings.ToUpper(keyword)) {
			return true
		}
	}

	return false
}

func (info *MediaInfo) processTechnicalTokens(tokens []string) {
	for _, token := range tokens {
		upper := strings.ToUpper(token)

		// Detectar calidad
		for _, quality := range qualityKeywords {
			if strings.Contains(upper, strings.ToUpper(quality)) {
				info.Quality = quality
				break
			}
		}

		// Detectar fuente
		for _, source := range sourceKeywords {
			if strings.Contains(upper, strings.ToUpper(source)) {
				info.Source = source
				break
			}
		}

		// Detectar idioma
		for _, lang := range languageKeywords {
			if strings.Contains(upper, strings.ToUpper(lang)) {
				info.Language = lang
				break
			}
		}
	}
}

func buildCleanTitle(tokens []string) string {
	var cleanTokens []string

	for _, token := range tokens {
		// Capitalizar primera letra de cada palabra
		if len(token) > 0 {
			cleanToken := strings.ToLower(token)
			cleanToken = strings.ToUpper(string(cleanToken[0])) + cleanToken[1:]
			cleanTokens = append(cleanTokens, cleanToken)
		}
	}

	return strings.Join(cleanTokens, " ")
}

// Funciones de utilidad para consultas API
func (info *MediaInfo) ToTMDBMovieQuery() string {
	query := info.Title
	if info.Year > 0 {
		query += fmt.Sprintf(" (%d)", info.Year)
	}
	return query
}

func (info *MediaInfo) ToTMDBTVQuery() string {
	return info.Title // Para series, TMDB generalmente no necesita año en la búsqueda
}

func (info *MediaInfo) String() string {
	mediaType := "Movie"
	if !info.IsMovie {
		mediaType = "TV Show"
	}

	result := fmt.Sprintf("Type: %s\nTitle: %s", mediaType, info.Title)

	if info.Year > 0 {
		result += fmt.Sprintf("\nYear: %d", info.Year)
	}

	if info.Season > 0 {
		result += fmt.Sprintf("\nSeason: %d", info.Season)
	}

	if info.Episode > 0 {
		result += fmt.Sprintf("\nEpisode: %d", info.Episode)
	}

	if info.Quality != "" {
		result += fmt.Sprintf("\nQuality: %s", info.Quality)
	}

	if info.Source != "" {
		result += fmt.Sprintf("\nSource: %s", info.Source)
	}

	if info.Language != "" {
		result += fmt.Sprintf("\nLanguage: %s", info.Language)
	}

	return result
}

func main() {
	// Ejemplos de uso
	testFiles := []string{
		"The.Matrix.1999.1080p.BluRay.x264-GROUP.mkv",
		"Breaking.Bad.S01E01.720p.HDTV.x264-CTU.mkv",
		"Avengers Endgame (2019) [1080p] [BluRay] [5.1] [YTS.MX].mp4",
		"Game.of.Thrones.S08E06.The.Iron.Throne.1080p.AMZN.WEB-DL.DDP5.1.H.264-GoT.mkv",
		"Parasite.2019.KOREAN.1080p.BluRay.H264.AAC-VXT.mp4",
		"The Mandalorian S02E08 1080p WEB-DL DD5.1 H264-CMRG.mkv",
	}

	for _, filename := range testFiles {
		fmt.Printf("=== Parsing: %s ===\n", filename)
		info := ParseMediaFilename(filename)
		fmt.Println(info)
		fmt.Printf("TMDB Query: %s\n", info.ToTMDBMovieQuery())
		fmt.Println()
	}
}
