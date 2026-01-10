package client

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

type Watcher struct {
	watchDir  string
	serverURL string
}

func NewWatcher(watchDir, serverURL string) *Watcher {
	return &Watcher{
		watchDir:  watchDir,
		serverURL: serverURL,
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
					log.Println("Modified file:", event.Name)
					// Now it can call UploadFile (capital U - exported function)
					if err := UploadFile(event.Name, w.serverURL); err != nil {
						log.Println("Error uploading:", err)
					}
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
