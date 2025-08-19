package transformations

import "strings"

type Cleaner struct {
	junkPatterns map[string]struct{}
}

func NewCleaner(patterns []string) *Cleaner {
	junk := make(map[string]struct{}, len(patterns))
	for _, p := range patterns {
		junk[strings.ToLower(p)] = struct{}{}
	}

	return &Cleaner{
		junkPatterns: junk,
	}
}

func (c *Cleaner) isJunk(token string) bool {
	_, exists := c.junkPatterns[strings.ToLower(token)]
	return exists
}

func (c *Cleaner) Clean(tokens []string) []string {
	var cleaned []string
	for _, token := range tokens {
		if !c.isJunk(token) {
			cleaned = append(cleaned, token)
		}
	}
	return cleaned
}
