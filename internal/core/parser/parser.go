package parser

import models "github.com/alejandro-bustamante/flick/internal/models"

// Tokenizer maneja la división del texto en tokens
type Tokenizer interface {
	Tokenize(input string) []string
}

// Cleaner limpia y normaliza tokens
type Cleaner interface {
	Clean(tokens []string) []string
}

// Extractor extrae información específica de tokens
type Extractor interface {
	Extract(tokens []string) *models.MediaInfo
}

// Logger para debugging y testing
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type Normalizer interface {
	NormalizeTokens(tokens []string) []string
}

type MediaParser struct {
	tokenizer  Tokenizer
	cleaner    Cleaner
	extractor  Extractor
	logger     Logger
	normalizer Normalizer
}

func NewMediaParser(t Tokenizer, c Cleaner, e Extractor, l Logger, n Normalizer) *MediaParser {
	return &MediaParser{
		tokenizer:  t,
		cleaner:    c,
		extractor:  e,
		logger:     l,
		normalizer: n,
	}
}

// The two Parse functions asume the <filename> recieved HAS NO EXTENSION
func (p *MediaParser) Parse(filename string) *models.ParseResult {
	result := &models.ParseResult{}

	p.logger.Debug("Parsing file: %s", filename)

	// Fase 1: Tokenización
	tokens := p.tokenizer.Tokenize(filename)
	p.logger.Debug("Tokens: %v", tokens)

	// Fase 2: Limpieza
	cleanTokens := p.cleaner.Clean(tokens)
	p.logger.Debug("Clean tokens: %v", cleanTokens)

	// Fase 3: Extracción
	info := p.extractor.Extract(cleanTokens)
	info.OriginalName = filename

	result.MediaInfo = info

	return result
}

func (p *MediaParser) ParseNormalized(filename string) *models.ParseResult {
	result := &models.ParseResult{}

	p.logger.Debug("Parsing file: %s", filename)

	// Stage 1: Tokenization
	tokens := p.tokenizer.Tokenize(filename)
	p.logger.Debug("Tokens: %v", tokens)

	// Stage 2: Normalization
	normalizedTokens := p.normalizer.NormalizeTokens(tokens)
	p.logger.Debug("Normalized tokens: %v", normalizedTokens)

	// Stage 3: Cleaning
	cleanTokens := p.cleaner.Clean(normalizedTokens)
	p.logger.Debug("Clean tokens: %v", cleanTokens)

	// Stage 4: Extraction
	info := p.extractor.Extract(cleanTokens)
	info.OriginalName = filename

	result.MediaInfo = info
	return result
}
