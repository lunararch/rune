package client

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watchDir  string
	serverURL string
	debounce  map[string]*time.Timer
	mu        sync.Mutex
}

func NewWatcher(watchDir, serverURL string) *Watcher {
	return &Watcher{
		watchDir:  watchDir,
		serverURL: serverURL,
		debounce:  make(map[string]*time.Timer),
	}
}

func (w *Watcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					w.handleFileChange(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	if err := watcher.Add(w.watchDir); err != nil {
		return err
	}

	log.Printf("Watching directory: %s\n", w.watchDir)
	<-done
	return nil
}

func (w *Watcher) handleFileChange(filename string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if timer, exists := w.debounce[filename]; exists {
		timer.Stop()
	}

	w.debounce[filename] = time.AfterFunc(500*time.Millisecond, func() {
		log.Printf("Uploading: %s\n", filepath.Base(filename))
		if err := UploadFile(filename, w.serverURL); err != nil {
			log.Printf("Error uploading %s: %v\n", filepath.Base(filename), err)
		} else {
			log.Printf("Successfully uploaded: %s\n", filepath.Base(filename))
		}

		w.mu.Lock()
		delete(w.debounce, filename)
		w.mu.Unlock()
	})
}
