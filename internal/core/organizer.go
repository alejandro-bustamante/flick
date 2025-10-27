package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alejandro-bustamante/flick/internal/models"
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
	watchDir  string
	moviesDir string
	seriesDir string
}

func NewOrganizer(p Parser, f Finder, watchDir, moviesDir, seriesDir string) *Organizer {
	return &Organizer{
		parser:    p,
		finder:    f,
		watchDir:  watchDir,
		moviesDir: moviesDir,
		seriesDir: seriesDir,
	}
}

// Equivalent to a dry run
func (o *Organizer) GetFinalDir(filePath string) (finalPath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		return ""
	}
	if info.IsDir() {
		// Should not be a directory
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
