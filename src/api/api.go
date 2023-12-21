package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type API struct {
	ListendAddr string
}

func New(listendAddr string) *API {
	return &API{
		ListendAddr: listendAddr,
	}
}

func (api *API) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (api *API) Start() {
	log.Println("Starting API server on", api.ListendAddr)
	log.Fatal(http.ListenAndServe(api.ListendAddr, nil))
}

func (api *API) Compare(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	f := []string{}

	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Ensure at least 2 files and at most 10 files are present
	files := r.MultipartForm.File["files"]
	if len(files) < 2 || len(files) > 10 {
		http.Error(w, "Invalid number of files", http.StatusBadRequest)
		return
	}

	// Create a folder named "files" if it doesn't exist
	err = os.MkdirAll("files", os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating files folder", http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer src.Close()

		dst, err := os.Create(fmt.Sprintf("files/%s", file.Filename))
		f = append(f, fmt.Sprintf("files/%s", file.Filename))

		if err != nil {
			http.Error(w, "Error creating destination file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			http.Error(w, "Error copying file", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	final, err := CompareFiles(f...)
	if err != nil {
		http.Error(w, "Error comparing files", http.StatusInternalServerError)
		return
	}

	// Return the changes as json
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&final)
}
