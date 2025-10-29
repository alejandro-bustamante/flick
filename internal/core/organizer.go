package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alejandro-bustamante/flick/internal/models"
	"github.com/alejandro-bustamante/flick/internal/watcher"
)

type Finder interface {
	GetMediaInfo(mediaInfo models.MediaInfo) (*models.MediaInfo, error)
}

type Parser interface {
	Parse(filename string) *models.ParseResult
	ParseNormalized(filename string) *models.ParseResult
}

type Organizer struct {
	parser    Parser
	finder    Finder
	watcher   *watcher.FolderWatcher
	watchDir  string
	moviesDir string
	seriesDir string
}

func NewOrganizer(p Parser, f Finder, w *watcher.FolderWatcher, moviesDir, seriesDir string) *Organizer {
	return &Organizer{
		parser:    p,
		finder:    f,
		watcher:   w,
		watchDir:  w.Config.Path,
		moviesDir: moviesDir,
		seriesDir: seriesDir,
	}
}

func (o *Organizer) Run() {
	if err := o.watcher.Start(); err != nil {
		log.Fatalf("Error starting watcher: %v", err)
	}

	log.Println("Organizer is running and listening for stable files...")

	go func() {
		for filePath := range o.watcher.StableFiles {
			log.Printf("Organizer received stable file: %s", filePath)

			// Process the file
			finalPath := o.GetFinalDir(filePath)
			if finalPath == "" {
				log.Printf("Could not determine final path for: %s", filePath)
				continue
			}

			log.Printf("Calculated final path: %s", finalPath)
		}

		log.Println("Watcher channel closed. Organizer stopping.")
	}()
}

// Equivalent to a dry run
func (o *Organizer) GetFinalDir(filePath string) (finalPath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		return ""
	}
	if info.IsDir() {
		return ""
	}

	fileName := filepath.Base(filePath)
	cleanFileName := o.parser.ParseNormalized(fileName)
	mediaInfo, err := o.finder.GetMediaInfo(*cleanFileName.MediaInfo)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println(mediaInfo.Title)
	fmt.Println(mediaInfo.Year)
	fmt.Printf("Accuracy: %v\n", mediaInfo.Accuracy)

	if mediaInfo.IsSeries {
		finalPath = o.seriesDir + mediaInfo.Title
	} else {
		folderName := mediaInfo.Title + "(" + strconv.Itoa(mediaInfo.Year) + ")"
		// E.G. /base/movies/directory/Titatnic(1999)/Titanic(1999).mkv
		finalPath = filepath.Join(o.moviesDir, folderName, folderName+filepath.Ext(fileName))
	}

	return finalPath
}
