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
	// --- Carga de configuración (sin cambios) ---
	data, err := config.LoadData("/home/alejandro/nvme/Repositorios/Developer/flick/patterns.toml")
	if err != nil {
		log.Fatalf("Error al cargar patterns.toml: %v", err)
	}

	sttgs, err := config.LoadSettings("/home/alejandro/nvme/Repositorios/Developer/flick/settings.toml")
	if err != nil {
		log.Fatalf("Error al cargar settings.toml: %v", err)
	}

	// --- Inicialización de componentes (sin cambios) ---
	tokenizer := transformations.NewTokenizer(data.Tokenizer.Separators)
	cleaner := transformations.NewCleaner(data.Cleaner.JunkPatterns)
	extrator := transformations.NewExtractor(data.Extractor.YearRange[:])
	logger := transformations.NewLogger("debug")
	nomalizer := transformations.NewNormalizer()

	p := parser.NewMediaParser(tokenizer, cleaner, extrator, logger, nomalizer)
	f := finder.NewTMDBFinder(sttgs.Secrets.TMDB_API_Key)
	o := core.NewOrganizer(p, f, sttgs.Directories.Watch, sttgs.Directories.Movies, sttgs.Directories.Series)

	// --- Lógica del Watcher (sin cambios) ---
	fileHandler := func(filePath string) {
		finalDir := o.GetFinalDir(filePath)
		// Usamos log en lugar de fmt para un output más estándar en daemons
		log.Printf("Archivo procesado: %s, Destino final: %s\n", filePath, finalDir)
	}

	watcherConfig := watcher.WatcherConfig{
		Path:           sttgs.Directories.Watch,
		StabilityDelay: 2 * time.Second,
		Recursive:      true,
	}

	folderWatcher, err := watcher.NewWatcher(watcherConfig, fileHandler)
	if err != nil {
		log.Fatalf("Error al crear el watcher: %v", err)
	}

	// Iniciamos el watcher para que se ejecute en segundo plano
	err = folderWatcher.Start()
	if err != nil {
		log.Fatalf("Error al iniciar el watcher: %v", err)
	}

	// --- Integración del Daemon ---
	// 1. Creamos una nueva instancia del daemon.
	flickDaemon, err := daemon.NewDaemon()
	if err != nil {
		log.Fatalf("No se pudo iniciar el daemon: %v", err)
	}

	// 2. Iniciamos el daemon. Esta función es bloqueante y se encargará
	//    de mantener la aplicación viva, reemplazando a select {}.
	//    También manejará el apagado seguro al recibir señales como Ctrl+C.
	flickDaemon.Start()

	log.Println("Flick se ha detenido.")
}
