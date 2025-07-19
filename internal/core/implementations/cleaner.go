package implementations

import "strings"

type Cleaner struct {
	junkPatterns []string
	minLength    int
}

func NewCleaner() *Cleaner {
	return &Cleaner{
		junkPatterns: []string{"xxx", "rarbg", "yts", "etrg", "yify"},
		minLength:    2,
	}
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

func (c *Cleaner) IsJunk(token string) bool {
	lower := strings.ToLower(token)

	for _, pattern := range c.junkPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Tokens muy cortos que no son años ni episodios
	if len(token) <= 1 {
		return true
	}

	return false
}
