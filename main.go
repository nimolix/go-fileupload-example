package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Streams upload directly from file -> mime/multipart -> pipe -> http-request
func streamingUploadFile(params map[string]string, paramName, path, boundary string, w *io.PipeWriter, file *os.File) {
	defer file.Close()
	defer w.Close()
	writer := multipart.NewWriter(w)
	writer.SetBoundary(boundary)

	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal(err)
		return
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	boundary := RandomBoundary()
	r, w := io.Pipe()
	go streamingUploadFile(params, paramName, path, boundary, w, file)
	request, err := http.NewRequest("POST", uri, r)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	return request, err
}

func RandomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func main() {
	path, _ := os.Getwd()
	path += "/test.pdf"
	extraParams := map[string]string{
		"title":       "My Document",
		"author":      "Matt Aimonetti",
		"description": "A document with all the Go programming language secrets",
	}
	request, err := newfileUploadRequest("http://localhost:8080", extraParams, "file", "test.pdf")
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		_, err := io.Copy(os.Stdout, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
	}
}
