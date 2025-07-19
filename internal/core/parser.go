package core

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
	Extract(tokens []string) *MediaInfo
}

// Validator valida resultados de parsing
type Validator interface {
	Validate(info *MediaInfo) error
	ValidateTitle(title string) error
}

// Logger para debugging y testing
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type MediaParser struct {
	tokenizer Tokenizer
	cleaner   Cleaner
	extractor Extractor
	validator Validator
	logger    Logger
}

func NewMediaParser(t Tokenizer, c Cleaner, e Extractor, v Validator, l Logger) *MediaParser {
	return &MediaParser{
		tokenizer: t,
		cleaner:   c,
		extractor: e,
		validator: v,
		logger:    l,
	}
}
func (p *MediaParser) Parse(filename string) *ParseResult {
	result := &ParseResult{}

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
func CleanMediaFile(filename string) string {
	return ""
}

func (p *MediaParser) calculateConfidence(info *MediaInfo) float32 {
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
