package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WatcherConfig struct {
	Path           string
	StabilityDelay time.Duration
	Recursive      bool
}

type FolderWatcher struct {
	config    WatcherConfig
	handler   func(string)
	watcher   *fsnotify.Watcher
	stopCh    chan bool
	pendingCh chan string
	done      chan bool
}

func NewWatcher(config WatcherConfig, handler func(string)) (*FolderWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating a watcher: %v", err)
	}

	fw := &FolderWatcher{
		config:    config,
		handler:   handler,
		watcher:   watcher,
		stopCh:    make(chan bool),
		pendingCh: make(chan string, 100),
		done:      make(chan bool),
	}

	return fw, nil
}

func (fw *FolderWatcher) Start() error {
	err := fw.watcher.Add(fw.config.Path)
	if err != nil {
		return fmt.Errorf("error adding the path %s: %v", fw.config.Path, err)
	}

	if fw.config.Recursive {
		filepath.Walk(fw.config.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() && path != fw.config.Path {
				fw.watcher.Add(path)
			}
			return nil
		})
	}

	go fw.watchFiles()
	go fw.processStability()

	log.Printf("Watcher iniciado para: %s", fw.config.Path)
	return nil
}

func (fw *FolderWatcher) Stop() {
	close(fw.stopCh)
	<-fw.done
	fw.watcher.Close()
	log.Println("Watcher stopped")
}

func (fw *FolderWatcher) watchFiles() {
	defer close(fw.done)

	for {
		select {
		case <-fw.stopCh:
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Only process created or modified files
			if event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Write == fsnotify.Write {

				// If its a directory and we're on recursive mode, add it to the watcher
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() && fw.config.Recursive {
					fw.watcher.Add(event.Name)
					continue
				}

				// If its a file, verify if we should process it
				if fw.shouldProcessFile(event.Name) {
					select {
					case fw.pendingCh <- event.Name:
					default:
						log.Printf("Pending files channel full, ignoring: %s", event.Name)
					}
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (fw *FolderWatcher) processStability() {
	pending := make(map[string]time.Time)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fw.stopCh:
			return

		case filePath := <-fw.pendingCh:
			pending[filePath] = time.Now()

		case <-ticker.C:
			now := time.Now()
			for filePath, addTime := range pending {
				if now.Sub(addTime) >= fw.config.StabilityDelay {
					if fw.isFileStable(filePath) {
						fw.handler(filePath)
						delete(pending, filePath)
					} else {
						// Reset the time if the file is still changing
						pending[filePath] = now
					}
				}
			}
		}
	}
}

func (fw *FolderWatcher) shouldProcessFile(filePath string) bool {
	// File extensions (temporal/incomplete) we should ignore
	ext := strings.ToLower(filepath.Ext(filePath))
	tempExtensions := []string{
		".tmp", ".part", ".crdownload", ".download", ".partial",
		".!qb", ".!ut", // BitTorrent
		".opdownload", // Opera
		".wkdownload", // Firefox temporal
		".filepart",   // Firefox
		".bc!",        // BitComet
		".dltemp",     // Download temporal
	}

	return !slices.Contains(tempExtensions, ext)
}

func (fw *FolderWatcher) isFileStable(filePath string) bool {
	info1, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	time.Sleep(500 * time.Millisecond)

	info2, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// The file is stable if it hasn't changed its size and modification time
	return info1.Size() == info2.Size() && info1.ModTime().Equal(info2.ModTime())
}
