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
	// Normalizar separadores
	normalized := input
	for _, sep := range t.separators {
		if sep != " " {
			normalized = strings.ReplaceAll(normalized, sep, " ")
		}
	}

	// Dividir por espacios
	tokens := strings.Fields(normalized)

	// Remover extensión del último token si existe
	if len(tokens) > 0 {
		lastToken := tokens[len(tokens)-1]
		if strings.Contains(lastToken, ".") {
			parts := strings.Split(lastToken, ".")
			if len(parts) > 1 {
				tokens[len(tokens)-1] = strings.Join(parts[:len(parts)-1], ".")
			}
		}
	}

	return tokens
}
