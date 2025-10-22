package transformations

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Normalizer struct {
}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// Removes case, diacritics, and non-alphanumeric characters.
func (n *Normalizer) NormalizeTokens(tokens []string) []string {
	// Compose a transformer that:
	// - Decomposes Unicode characters (NFD),
	// - Removes non-spacing marks (diacritics),
	// - Recomposes characters to NFC form.
	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)

	// Apply the chained transformer to the input string.
	// Returns transformed text, number of bytes read, and error.
	for i, token := range tokens {
		text, _, _ := transform.String(t, token)
		tokens[i] = strings.ToLower(text)
	}
	return tokens
}
