package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/joho/godotenv/autoload"
)

var serverPort string
var storageDir string

func init() {
	serverPort = os.Getenv("PORT")
	storageDir = os.Getenv("STORAGEDIR")
}

func main() {
	if err := os.MkdirAll(storageDir, 0775); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download/", handleDownload)
	http.HandleFunc("/list", handleList)

	fmt.Printf("Server starting on port %s...\n", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create(filepath.Join(storageDir, header.Filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File %s uploaded successfully\n", header.Filename)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	filePath := filepath.Join(storageDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}

func handleList(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(storageDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "[")
	for i, file := range files {
		if i > 0 {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `"%s"`, file.Name())
	}
	fmt.Fprintf(w, "]")
}
