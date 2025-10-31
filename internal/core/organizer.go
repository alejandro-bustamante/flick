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
			mediaDirs := o.GetFinalDir(filePath)
			if mediaDirs.DestinyPath == "" {
				log.Printf("Could not determine final path for: %s", filePath)
				continue
			}

			// Permisions that allows to read and write for any user
			dirPerm := 0777
			os.MkdirAll(mediaDirs.DestinyDir, os.FileMode(dirPerm))
			err := os.Rename(mediaDirs.OriginalPath, mediaDirs.DestinyPath)
			if err != nil {
				log.Printf("Could not move to final path. Error: %s", err)
			}

			log.Printf("Calculated final path: %s", mediaDirs.DestinyPath)
		}

		log.Println("Watcher channel closed. Organizer stopping.")
	}()
}

// Equivalent to a dry run
func (o *Organizer) GetFinalFilePath(filePath string) (finalPath string) {
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

func (o *Organizer) GetFinalDir(filePath string) models.MediaPaths {
	info, err := os.Stat(filePath)
	if err != nil {
		return models.MediaPaths{}
	}
	if info.IsDir() {
		return models.MediaPaths{}

	}

	fileName := filepath.Base(filePath)
	cleanFileName := o.parser.ParseNormalized(fileName)
	mediaInfo, err := o.finder.GetMediaInfo(*cleanFileName.MediaInfo)
	if err != nil {
		fmt.Println(err)
		return models.MediaPaths{}
	}
	fmt.Println(mediaInfo.Title)
	fmt.Println(mediaInfo.Year)
	fmt.Printf("Accuracy: %v\n", mediaInfo.Accuracy)

	var mediaPaths models.MediaPaths
	mediaPaths.OriginalPath = filePath
	mediaPaths.OriginalDir = filepath.Dir(filePath)

	if mediaInfo.IsSeries {
		mediaPaths.DestinyDir = o.moviesDir
		mediaPaths.DestinyPath = o.moviesDir + mediaInfo.Title
		// finalPath = o.seriesDir + mediaInfo.Title
	} else {
		folderName := mediaInfo.Title + "(" + strconv.Itoa(mediaInfo.Year) + ")"
		// E.G. /base/movies/directory/Titatnic(1999)/Titanic(1999).mkv
		mediaPaths.DestinyDir = filepath.Join(o.moviesDir, folderName)
		mediaPaths.DestinyPath = filepath.Join(mediaPaths.DestinyDir, folderName+filepath.Ext(fileName))
	}

	return mediaPaths
}
