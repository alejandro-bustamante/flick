// Strategy
// Archivo → Normalización → Tokenización → Limpieza → Extracción → Resultado
package main

import (
	"fmt"

	parser "github.com/alejandro-bustamante/flick/internal/core"
	implementations "github.com/alejandro-bustamante/flick/internal/core/implementations"
)

func main() {
	tokenizer := implementations.NewTokenizer()
	cleaner := implementations.NewCleaner()
	extrator := implementations.NewExtractor()
	validator := implementations.NewValidator()
	logger := implementations.NewLogger("info")
	nomalizer := implementations.NewNormalizer()

	p := parser.NewMediaParser(tokenizer, cleaner, extrator, validator, logger, nomalizer)
	result := p.Parse("The.Death.Of.Superman.2.2018.720p.WEBRip.x264-[YTS.AM].mp4")
	fmt.Println(result.MediaInfo.Title)
	result = p.ParseNormalized("Sangre.de.Cóndor.2018.720p.WEBRip.x264-[YTS.AM].mp4")
	fmt.Println(result.MediaInfo.Title)
}
