package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/joho/godotenv/autoload"
)

var serverUrl string

func init() {
	serverUrl = os.Getenv("SERVERURL") + ":" + os.Getenv("PORT")
}

func main() {
	if len(os.Args) > 3 {
		fmt.Println("Usage: client <upload|download> <filename>")
		os.Exit(1)
	}

	command := os.Args[1]
	filename := os.Args[2]

	switch command {
	case "upload":
		if err := uploadFile(filename); err != nil {
			fmt.Println("Error uploading", err)
			return
		}
		fmt.Println("Upload Successful")
	case "download":
		if err := downloadFile(filename); err != nil {
			fmt.Println("Error downloading", err)
			return
		}
		fmt.Println("Download Successful")
	default:
		fmt.Println("Unknown command:", command)

	}
}

func uploadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", serverUrl+"/upload", body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil

}

func downloadFile(filename string) error {
	resp, err := http.Get(serverUrl + "/download/" + filename)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
