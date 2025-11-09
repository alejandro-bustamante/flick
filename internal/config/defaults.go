package config

import (
	_ "embed"
)

//go:embed patterns.toml
var DefaultPatterns []byte

// defaultSettings is a template for the settings.toml file.
// We use a string template because it requires user-specific secrets.
const DefaultSettings = `[directories]
# Path to the folder Flick should monitor for new files.
# watch = "/path/to/your/downloads"
watch = ""

# Path where organized movie files will be moved.
# movies = "/path/to/your/movies"
movies = ""

# Path where organized TV show files will be moved.
# series = "/path/to/your/tvshows"
series = ""

[secrets]
# Your API key from The Movie Database (TMDb).
# This is required to fetch media information.
# Get one here: https://www.themoviedb.org/signup
# tmdb_api_key = "YOUR_API_KEY_HERE"
tmdb_api_key = ""
`
