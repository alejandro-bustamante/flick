package finders

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	models "github.com/alejandro-bustamante/flick/internal/models"
)

type SearchResponse struct {
	Page         int            `json:"page"`
	Results      []SearchResult `json:"results"`
	TotalPages   int            `json:"total_pages"`
	TotalResults int            `json:"total_results"`
}

type SearchResult struct {
	ID int `json:"id"`
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

func (f *TMDBFinder) GetMediaInfo(mediaInfo models.MediaInfo) (*models.MediaInfo, error) {
	movieID, err := f.searchForID(mediaInfo)
	if err != nil {
		return nil, err
	}

	if movieID == 0 {
		return nil, fmt.Errorf("no se encontraron resultados para: %s", mediaInfo.Title)
	}

	title, year, err := f.getDetailsByID(movieID, mediaInfo.IsSeries)
	if err != nil {
		return nil, err
	}

	year_int, _ := strconv.Atoi(year)
	return &models.MediaInfo{
		Title: title,
		Year:  year_int,
	}, nil
}

func (f *TMDBFinder) searchForID(mediaInfo models.MediaInfo) (int, error) {
	var baseURL string
	if mediaInfo.IsSeries {
		baseURL = "https://api.themoviedb.org/3/search/tv"
	} else {
		baseURL = "https://api.themoviedb.org/3/search/movie"
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return 0, err
	}

	q := u.Query()
	q.Set("include_adult", "true")
	q.Set("language", "en-US")
	q.Set("page", "1")
	q.Set("query", mediaInfo.Title)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+f.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return 0, err
	}

	if len(searchResponse.Results) == 0 {
		return 0, nil
	}

	return searchResponse.Results[0].ID, nil
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
