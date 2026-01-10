package main

import (
	"fmt"
	"os"

	"github.com/lunararch/rune/internal/client"

	_ "github.com/joho/godotenv/autoload"
)

var serverUrl string

func init() {
	serverUrl = os.Getenv("SERVERURL") + ":" + os.Getenv("PORT")
}

func main() {
	if len(os.Args) > 3 {
		fmt.Println("Usage: client <upload|download> <path>")
		os.Exit(1)
	}

	command := os.Args[1]
	path := os.Args[2]

	switch command {
	case "upload":
		if err := client.UploadFile(path, serverUrl); err != nil {
			fmt.Println("Error uploading", err)
			return
		}
		fmt.Println("Upload Successful")
	case "download":
		if err := client.DownloadFile(path, serverUrl); err != nil {
			fmt.Println("Error downloading", err)
			return
		}
		fmt.Println("Download Successful")
	case "watch":
		watcher := client.NewWatcher(path, serverUrl)
		if err := watcher.Start(); err != nil {
			fmt.Println("Error starting watcher:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown command:", command)

	}
}
