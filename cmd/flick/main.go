// Strategy
// Archivo → Normalización → Tokenización → Limpieza → Extracción → Resultado
package main

import (
	"fmt"

	config "github.com/alejandro-bustamante/flick/internal/config"
	"github.com/alejandro-bustamante/flick/internal/core"
	finder "github.com/alejandro-bustamante/flick/internal/core/finder"
	parser "github.com/alejandro-bustamante/flick/internal/core/parser"
	"github.com/alejandro-bustamante/flick/internal/core/parser/transformations"
)

func main() {
	data, err := config.LoadData("patterns.toml")
	if err != nil {
		panic(err)
	}

	stgs, err := config.LoadSettings("settings.toml")
	if err != nil {
		panic(err)
	}

	tokenizer := transformations.NewTokenizer(data.Tokenizer.Separators)
	cleaner := transformations.NewCleaner(data.Cleaner.JunkPatterns)
	extrator := transformations.NewExtractor(data.Extractor.YearRange[:])
	logger := transformations.NewLogger("info")
	nomalizer := transformations.NewNormalizer()

	p := parser.NewMediaParser(tokenizer, cleaner, extrator, logger, nomalizer)
	result := p.Parse("The-Shawshank-redemption.2019.720p.WEBRip.x264-[YTS.AM].mp4")
	fmt.Println(result.MediaInfo.Title)

	f := finder.NewTMDBFinder(stgs.Secrets.TMDB_API_Key)
	o := core.NewOrganizer(p, f, stgs.Directories.Watch, stgs.Directories.Movies, stgs.Directories.Series)
	fmt.Println(o.GetFinalDir("The-Shawshank-redemption.2019.720p.WEBRip.x264-[YTS.AM].mp4"))

}
