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
	// data, err := config.LoadData("patterns.toml")
	data, err := config.LoadData("/home/alejandro/nvme/Repositorios/Developer/flick/patterns.toml")

	if err != nil {
		panic(err)
	}

	// sttgs, err := config.LoadSettings("settings.toml")
	sttgs, err := config.LoadSettings("/home/alejandro/nvme/Repositorios/Developer/flick/settings.toml")
	if err != nil {
		panic(err)
	}

	tokenizer := transformations.NewTokenizer(data.Tokenizer.Separators)
	cleaner := transformations.NewCleaner(data.Cleaner.JunkPatterns)
	extrator := transformations.NewExtractor(data.Extractor.YearRange[:])
	logger := transformations.NewLogger("debug")
	nomalizer := transformations.NewNormalizer()

	p := parser.NewMediaParser(tokenizer, cleaner, extrator, logger, nomalizer)

	f := finder.NewTMDBFinder(sttgs.Secrets.TMDB_API_Key)
	o := core.NewOrganizer(p, f, sttgs.Directories.Watch, sttgs.Directories.Movies, sttgs.Directories.Series)
	fmt.Println(o.GetFinalDir("/home/alejandro/nvme/Repositorios/Developer/test_flick/Titanic(1999).720p.WEBRip.x264-[YTS.AM].mp4"))

}
