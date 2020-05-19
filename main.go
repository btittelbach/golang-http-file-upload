package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadSize = 8 * 1024 * 1024 // 8 mb
var servePath string
var bindipport string

func environOrDefault(envvarname, defvalue string) string {
	if len(os.Getenv(envvarname)) > 0 {
		return os.Getenv(envvarname)
	} else {
		return defvalue
	}
}

func main() {
	uploadPath1 := environOrDefault("GOLANGHTTPUPLOAD_UPLOAD_PATH1", "./uploadsA4")
	uploadPath2 := environOrDefault("GOLANGHTTPUPLOAD_UPLOAD_PATH2", "./uploadsA3")
	servePath = environOrDefault("GOLANGHTTPUPLOAD_SERVE_PATH", "./public")
	bindipport = environOrDefault("GOLANGHTTPUPLOAD_BINDIP_PORT", ":8080")

	upload1Handler := uploadFileHandler(uploadPath1)
	upload2Handler := uploadFileHandler(uploadPath2)

	http.HandleFunc("/upload", upload1Handler)
	http.HandleFunc("/uploadA4", upload1Handler)
	http.HandleFunc("/uploadA3", upload2Handler)

	fs := http.FileServer(http.Dir(servePath))
	http.Handle("/", http.StripPrefix("/", fs))

	log.Print("Server started on ", bindipport, ", use /upload for uploading files")
	log.Fatal(http.ListenAndServe(bindipport, nil))
}

func uploadFileHandler(uploadPath string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		file, _, err := r.FormFile("file")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		detectedFileType := http.DetectContentType(fileBytes)
		switch detectedFileType {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
		case "application/postscript":
		case "text/plain":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}
		fileName := randToken(12)
		fileEndings, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join(uploadPath, "."+fileName+fileEndings[0])
		finalPath := filepath.Join(uploadPath, fileName+fileEndings[0])
		fmt.Printf("FileType: %s, File: %s\n", detectedFileType, finalPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			os.Remove(newPath)
			return
		}
		if os.Rename(newPath, finalPath) != nil {
			w.Write([]byte("MV ERROR"))
			os.Remove(newPath)
			return
		}
		w.Write([]byte("SUCCESS"))
	})
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
