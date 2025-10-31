package models

type MediaInfo struct {
	Title        string
	IsSeries     bool
	Year         int
	Season       int
	Episode      int
	Remaining    []string
	OriginalName string // For debugging
	Accuracy     int    // (0-5)
}

type ParseResult struct {
	MediaInfo *MediaInfo
	Errors    []error
	Warnings  []string
}

type TestCase struct {
	Input       string
	Expected    *MediaInfo
	ShouldError bool
}

type UserSettings struct {
	Directories struct {
		Watch  string `toml:"watch"`
		Movies string `toml:"movies"`
		Series string `toml:"series"`
	} `toml:"directories"`
	Secrets struct {
		TMDB_API_Key string `toml:"tmdb_api_key"`
	} `toml:"secrets"`
}

type Config struct {
	Tokenizer struct {
		Separators []string `toml:"separators"`
	} `toml:"tokenizer"`

	Cleaner struct {
		JunkPatterns []string `toml:"junk_patterns"`
		MinLength    int      `toml:"min_length"`
	} `toml:"cleaner"`
	Extractor struct {
		YearRange [2]int `toml:"year_range"`
	} `toml:"extractor"`
}
