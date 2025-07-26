// DefaultTokenizer - Implementación básica del tokenizer
package implementations

import "strings"

type Tokenizer struct {
	separators []string
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		separators: []string{".", "-", "_", " ", "[", "]", "(", ")", "{", "}"},
	}
}

func (t *Tokenizer) Tokenize(input string) []string {
	normalized := input
	for _, sep := range t.separators {
		if sep != " " {
			normalized = strings.ReplaceAll(normalized, sep, " ")
		}
	}
	tokens := strings.Fields(normalized)
	return tokens
}
