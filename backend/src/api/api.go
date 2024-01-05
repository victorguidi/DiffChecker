package api

import (
	"docfiff/src/db"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
)

var validate = validator.New()

type API struct {
	ListendAddr string
	db          *db.DB
}

func New(listendAddr string) *API {
	d, err := db.NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	return &API{
		ListendAddr: listendAddr,
		db:          d,
	}
}

func (api *API) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (api *API) Start() {
	log.Println("Starting API server on", api.ListendAddr)
	log.Fatal(http.ListenAndServe(api.ListendAddr, nil))
}

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Content-Type", "application/json")
}

func (api *API) GetDiffs(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.URL.Query().Get("email")
	var query bson.D
	if email != "" {
		type Query struct {
			Author string `validate:"email"`
		}
		if err := validate.Struct(Query{Author: email}); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query = bson.D{{"author", email}}
	} else {
		query = bson.D{}
	}

	var diffs []Response
	err := api.db.FindAll(&diffs, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if diffs == nil {
		diffs = []Response{}
	}
	json.NewEncoder(w).Encode(diffs)
}

func (api *API) GetDiffBy(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	type RQuery struct {
		Id string `validate:"uuid"`
	}

	if err := validate.Struct(RQuery{Id: id}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var diff Response

	err := api.db.FindDiffBy(bson.D{{"id", id}}, &diff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(diff)

}

func (api *API) Compare(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
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
	email := r.MultipartForm.Value["email"]

	type Query struct {
		Author string `validate:"email"`
	}
	if err := validate.Struct(Query{Author: email[0]}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if (len(files) < 2 || len(files) > 10) || len(files)%2 != 0 {
		http.Error(w, "Please upload 2 to 10 files", http.StatusBadRequest)
		return
	}
	response, err := SaveFiles(files)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		type Error struct {
			Error string `json:"error"`
			Tip   string `json:"tip"`
		}
		e := Error{
			Error: err.Error(),
			Tip:   "Maybe you send two files with the same name?",
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	for _, r := range response {
		if err := r.CompareTwoFilesInDir(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Author = email[0]
		go api.db.InsertDiff(r)
	}

	json.NewEncoder(w).Encode(response)
}
