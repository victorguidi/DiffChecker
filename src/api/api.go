package api

import (
	"encoding/json"
	"log"
	"net/http"
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

	err := r.ParseMultipartForm(500 << 20) // 500 MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["files"]
	if (len(files) < 2 || len(files) > 10) || len(files)%2 != 0 {
		http.Error(w, "Please upload 2 to 10 files", http.StatusBadRequest)
		return
	}
	response, err := SaveFiles(files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	for _, r := range response {
		r.CompareTwoFilesInDir()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
