package main

import (
	"log"
	"time"

	config "github.com/alejandro-bustamante/flick/internal/config"
	"github.com/alejandro-bustamante/flick/internal/core"
	finder "github.com/alejandro-bustamante/flick/internal/core/finder"
	parser "github.com/alejandro-bustamante/flick/internal/core/parser"
	"github.com/alejandro-bustamante/flick/internal/core/parser/transformations"
	"github.com/alejandro-bustamante/flick/internal/daemon"
	"github.com/alejandro-bustamante/flick/internal/watcher"
)

func main() {
	// --- Load config (sin cambios) ---
	data, err := config.LoadData("/home/alejandro/nvme/Repositorios/Developer/flick/patterns.toml")
	if err != nil {
		log.Fatalf("Error al cargar patterns.toml: %v", err)
	}

	sttgs, err := config.LoadSettings("/home/alejandro/nvme/Repositorios/Developer/flick/settings.toml")
	if err != nil {
		log.Fatalf("Error al cargar settings.toml: %v", err)
	}

	// --- Initialize components ---
	tokenizer := transformations.NewTokenizer(data.Tokenizer.Separators)
	cleaner := transformations.NewCleaner(data.Cleaner.JunkPatterns)
	extrator := transformations.NewExtractor(data.Extractor.YearRange[:])
	logger := transformations.NewLogger("debug")
	nomalizer := transformations.NewNormalizer()

	p := parser.NewMediaParser(tokenizer, cleaner, extrator, logger, nomalizer)
	f := finder.NewTMDBFinder(sttgs.Secrets.TMDB_API_Key)
	o := core.NewOrganizer(p, f, sttgs.Directories.Watch, sttgs.Directories.Movies, sttgs.Directories.Series)

	watcherConfig := watcher.WatcherConfig{
		Path:           sttgs.Directories.Watch,
		StabilityDelay: 2 * time.Second,
		Recursive:      true,
	}

	folderWatcher, err := watcher.NewWatcher(watcherConfig, o)
	if err != nil {
		log.Fatalf("Error al crear el watcher: %v", err)
	}

	// --- Start watcher ---
	err = folderWatcher.Start()
	if err != nil {
		log.Fatalf("Error al iniciar el watcher: %v", err)
	}

	// --- Start the daemon (blocking function) ---
	flickDaemon, err := daemon.NewDaemon()
	if err != nil {
		log.Fatalf("No se pudo iniciar el daemon: %v", err)
	}

	flickDaemon.Start()

	log.Println("Flick se ha detenido.")
}
