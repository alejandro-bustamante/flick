package implementations

import "strings"

type Cleaner struct {
	junkPatterns map[string]struct{}
	minLength    int
}

func NewCleaner() *Cleaner {
	patterns := []string{
		"xxx", "rarbg", "yts", "etrg", "yify",
		"720p", "1080p", "4k", "2160p", "480p", "bluray", "brrip",
		"web-dl", "webrip", "hdrip", "hdtv", "netflix", "amzn",
		"spanish", "latino", "english", "dual",
	}
	junk := make(map[string]struct{}, len(patterns))
	for _, p := range patterns {
		junk[strings.ToLower(p)] = struct{}{}
	}

	return &Cleaner{
		junkPatterns: junk,
		minLength:    1,
	}
}

func (c *Cleaner) IsJunk(token string) bool {
	_, exists := c.junkPatterns[strings.ToLower(token)]
	return exists
}

func (c *Cleaner) Clean(tokens []string) []string {
	var cleaned []string
	for _, token := range tokens {
		if !c.IsJunk(token) && len(token) >= c.minLength {
			cleaned = append(cleaned, token)
		}
	}
	return cleaned
}
