package core

import (
	"fmt"
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
func (o *Organizer) GetFinalDir(filename string) (filePath string) {
	cleanFileName := o.parser.ParseNormalized(filename)
	mediaInfo, _ := o.finder.GetMediaInfo(*cleanFileName.MediaInfo)
	fmt.Println(mediaInfo.Title)
	fmt.Println(mediaInfo.Year)

	var finalPath, folderName string
	if mediaInfo.IsSeries {
		finalPath = o.seriesDir + folderName
	} else {
		folderName := mediaInfo.Title + strconv.Itoa(mediaInfo.Year)
		finalPath = filepath.Join(o.moviesDir, folderName, folderName+filepath.Ext(filename))
	}

	return finalPath
}
