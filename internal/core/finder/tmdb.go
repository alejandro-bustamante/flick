package finders

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode" // Importamos unicode

	models "github.com/alejandro-bustamante/flick/internal/models"
	"golang.org/x/text/runes"        // Importamos runes
	"golang.org/x/text/transform"    // Importamos transform
	"golang.org/x/text/unicode/norm" // Importamos norm
)

type SearchResponse struct {
	Page         int            `json:"page"`
	Results      []SearchResult `json:"results"`
	TotalPages   int            `json:"total_pages"`
	TotalResults int            `json:"total_results"`
}

type SearchResult struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`          // movies
	ReleaseDate  string `json:"release_date"`   // movies
	Name         string `json:"name"`           // series
	FirstAirDate string `json:"first_air_date"` // series
}

type MovieDetailsResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
}

type TVDetailsResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	FirstAirDate string `json:"first_air_date"`
}

type TMDBFinder struct {
	APIKey string
}

func NewTMDBFinder(APIKey string) *TMDBFinder {
	return &TMDBFinder{
		APIKey: APIKey,
	}
}

// remove diacritics and converts to lower case
func normalizeString(input string) string {
	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
	normalized, _, _ := transform.String(t, input)
	return strings.ToLower(normalized)
}

func getYearDigitDiff(year1 int, year2Str string) int {
	if year1 == 0 || year2Str == "" {
		return 99 // Missing year. Return a very high difference
	}
	year2, err := strconv.Atoi(year2Str)
	if err != nil {
		return 99
	}

	if year1 < 1000 || year1 > 9999 || year2 < 1000 || year2 > 9999 {
		return 99
	}

	s1 := strconv.Itoa(year1)
	s2 := strconv.Itoa(year2)

	diffCount := 0
	for i := range 4 {
		if s1[i] != s2[i] {
			diffCount++
		}
	}
	return diffCount
}

func (f *TMDBFinder) GetMediaInfo(mediaInfo models.MediaInfo) (*models.MediaInfo, error) {
	bestID, certainty, err := f.searchForID(mediaInfo)
	if err != nil {
		return nil, err
	}

	if bestID == 0 {
		return nil, fmt.Errorf("could not find results for: %s", mediaInfo.Title)
	}

	title, year, err := f.getDetailsByID(bestID, mediaInfo.IsSeries)
	if err != nil {
		return nil, err
	}

	year_int, _ := strconv.Atoi(year)

	return &models.MediaInfo{
		Title:    title,
		Year:     year_int,
		IsSeries: mediaInfo.IsSeries,
		Season:   mediaInfo.Season,
		Episode:  mediaInfo.Episode,
		Accuracy: certainty,
	}, nil
}

func (f *TMDBFinder) searchForID(mediaInfo models.MediaInfo) (bestID int, accuracy int, err error) {
	var baseURL string
	if mediaInfo.IsSeries {
		baseURL = "https://api.themoviedb.org/3/search/tv"
	} else {
		baseURL = "https://api.themoviedb.org/3/search/movie"
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return 0, 0, err
	}

	q := u.Query()
	q.Set("include_adult", "true")
	q.Set("language", "en-US")
	q.Set("page", "1")
	q.Set("query", mediaInfo.Title)

	if mediaInfo.Year > 0 {
		if mediaInfo.IsSeries {
			q.Set("first_air_date_year", strconv.Itoa(mediaInfo.Year))
		} else {
			q.Set("year", strconv.Itoa(mediaInfo.Year))
		}
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+f.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, 0, err
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return 0, 0, err
	}

	if len(searchResponse.Results) == 0 {
		return 0, 0, nil
	}

	// --- Match accuracy logic ---
	if mediaInfo.IsSeries {
		return searchResponse.Results[0].ID, 0, nil
	}

	bestID = searchResponse.Results[0].ID
	bestCertainty := 1

	normalizedLocalTitle := normalizeString(mediaInfo.Title)

	for _, result := range searchResponse.Results {
		currentCertainty := 0
		normalizedApiTitle := normalizeString(result.Title)

		apiYearStr := ""
		if result.ReleaseDate != "" {
			parts := strings.Split(result.ReleaseDate, "-")
			if len(parts) > 0 {
				apiYearStr = parts[0]
			}
		}
		yearDiff := getYearDigitDiff(mediaInfo.Year, apiYearStr)

		partialTitleMatch := strings.Contains(normalizedLocalTitle, normalizedApiTitle) || strings.Contains(normalizedApiTitle, normalizedLocalTitle)
		perfectTitleMatch := normalizedLocalTitle == normalizedApiTitle

		if perfectTitleMatch {
			if mediaInfo.Year == 0 {
				currentCertainty = 4 // perfect title, release year unknown
			} else if yearDiff == 0 {
				currentCertainty = 5 // perfect title, perfect year
			} else if yearDiff == 1 {
				currentCertainty = 3 // perfect title, type in year, error by one digit
			} else { // yearDiff >= 2
				currentCertainty = 2 // perfect title, 2+ digit error
			}
		} else if partialTitleMatch {
			if mediaInfo.Year == 0 {
				currentCertainty = 2 // parcial title, release year unknown
			} else if yearDiff == 0 {
				currentCertainty = 3 // parcial title, perfect year
			} else if yearDiff == 1 {
				currentCertainty = 2 // parcial title, 1 digit error
			} else { // yearDiff >= 2
				currentCertainty = 2 // parcial title, 2+ digit error
			}
		} else {
			currentCertainty = 1 // No partial or total matches
		}

		if currentCertainty > bestCertainty {
			bestCertainty = currentCertainty
			bestID = result.ID
		}

		if bestCertainty == 5 {
			break
		}
	}

	return bestID, bestCertainty, nil
}

func (f *TMDBFinder) getDetailsByID(id int, isSeries bool) (string, string, error) {
	var baseURL string
	if isSeries {
		baseURL = "https://api.themoviedb.org/3/tv/"
	} else {
		baseURL = "https://api.themoviedb.org/3/movie/"
	}

	url := baseURL + strconv.Itoa(id) + "?language=en-US"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+f.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	if isSeries {
		var tvDetails TVDetailsResponse
		if err := json.Unmarshal(body, &tvDetails); err != nil {
			return "", "", err
		}

		year := ""
		if tvDetails.FirstAirDate != "" {
			// Extract the year on format "YYYY-MM-DD"
			dateParts := strings.Split(tvDetails.FirstAirDate, "-")
			if len(dateParts) > 0 {
				year = dateParts[0]
			}
		}

		return tvDetails.Name, year, nil
	} else {
		var movieDetails MovieDetailsResponse
		if err := json.Unmarshal(body, &movieDetails); err != nil {
			return "", "", err
		}

		year := ""
		if movieDetails.ReleaseDate != "" {
			// Extract the year on format "YYYY-MM-DD"
			dateParts := strings.Split(movieDetails.ReleaseDate, "-")
			if len(dateParts) > 0 {
				year = dateParts[0]
			}
		}

		return movieDetails.Title, year, nil
	}
}
