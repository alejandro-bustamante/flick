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
	NormalizeForComparison(input string) string
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
			destinationPath := o.GetDestinationPath(filePath)
			destinationDir := filepath.Dir(destinationPath)
			if destinationPath == "" {
				log.Printf("Could not determine final path for: %s", filePath)
				continue
			}

			// Permisions that allows to read and write for any user
			dirPerm := 0777
			os.MkdirAll(destinationDir, os.FileMode(dirPerm))
			err := os.Rename(filePath, destinationPath)
			if err != nil {
				log.Printf("Could not move to final path. Error: %s", err)
			}

			log.Printf("Calculated final path: %s", destinationPath)
		}

		log.Println("Watcher channel closed. Organizer stopping.")
	}()
}

// Equivalent to a dry run
func (o *Organizer) GetDestinationPath(filePath string) string {
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

	var destinationPath string
	if mediaInfo.IsSeries {
		destinationPath = o.seriesDir + mediaInfo.Title
	} else {
		folderName := mediaInfo.Title + "(" + strconv.Itoa(mediaInfo.Year) + ")"
		// E.G. /base/movies/directory/Titatnic(1999)/Titanic(1999).mkv
		destinationPath = filepath.Join(o.moviesDir, folderName, folderName+filepath.Ext(fileName))
	}

	return destinationPath
}
