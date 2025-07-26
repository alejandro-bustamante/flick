package core

import models "github.com/alejandro-bustamante/flick/internal/models"

// FileProvider abstrae la fuente de nombres de archivos
type FileProvider interface {
	GetFiles() ([]string, error)
	GetFileByName(name string) (string, error)
}

// Tokenizer maneja la división del texto en tokens
type Tokenizer interface {
	Tokenize(input string) []string
}

// Cleaner limpia y normaliza tokens
type Cleaner interface {
	Clean(tokens []string) []string
	IsJunk(token string) bool
}

// Extractor extrae información específica de tokens
type Extractor interface {
	Extract(tokens []string) *models.MediaInfo
}

// Validator valida resultados de parsing
type Validator interface {
	Validate(info *models.MediaInfo) error
	ValidateTitle(title string) error
}

// Logger para debugging y testing
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type Normalizer interface {
	NormalizeTokens(tokens []string) []string
}

type MediaParser struct {
	tokenizer  Tokenizer
	cleaner    Cleaner
	extractor  Extractor
	validator  Validator
	logger     Logger
	normalizer Normalizer
}

func NewMediaParser(t Tokenizer, c Cleaner, e Extractor, v Validator, l Logger, n Normalizer) *MediaParser {
	return &MediaParser{
		tokenizer:  t,
		cleaner:    c,
		extractor:  e,
		validator:  v,
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

	// Fase 4: Validación
	if err := p.validator.Validate(info); err != nil {
		result.Errors = append(result.Errors, err)
	}

	result.MediaInfo = info
	result.Confidence = p.calculateConfidence(info)

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

	// Stage 5: Validation
	if err := p.validator.Validate(info); err != nil {
		result.Errors = append(result.Errors, err)
	}

	result.MediaInfo = info
	result.Confidence = p.calculateConfidence(info)

	return result
}

func (p *MediaParser) calculateConfidence(info *models.MediaInfo) float32 {
	confidence := float32(0.5) // Base confidence

	if info.Title != "" {
		confidence += 0.2
	}
	if info.Year > 0 {
		confidence += 0.2
	}
	if info.IsMovie || (info.Season > 0 && info.Episode > 0) {
		confidence += 0.1
	}

	return confidence
}
