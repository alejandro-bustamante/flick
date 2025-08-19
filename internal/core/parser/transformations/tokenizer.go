package transformations

import "strings"

type Tokenizer struct {
	separators []string
}

func NewTokenizer(separators []string) *Tokenizer {
	return &Tokenizer{separators: separators}
}

func (t *Tokenizer) Tokenize(input string) []string {
	for _, sep := range t.separators {
		input = strings.ReplaceAll(input, sep, " ")
	}
	tokens := strings.Fields(input)
	return tokens
}
