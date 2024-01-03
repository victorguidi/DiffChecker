package api

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type Response struct {
	Id      string   `json:"id"`
	Files   []string `json:"files"`
	Date    string   `json:"date"`
	Changes []Change `json:"changes"`
}

type Changes struct {
	Original   Change `json:"original"`
	Difference Change `json:"difference"`
}

type Change struct {
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// This function get all the files in the API and save them in the directory for use later
func SaveFiles(files []*multipart.FileHeader) ([]Response, error) {
	// Get the files from the API
	var responses []Response
	for i, file := range files {
		var response Response
		if i%2 == 0 {
			response.Id = uuid.New().String()
			response.Files = append(response.Files, file.Filename)
			response.Date = time.Now().String()
			responses = append(responses, response)
		} else {
			responses[len(responses)-1].Files = append(responses[len(responses)-1].Files, file.Filename)
		}
	}

	// The Directory for saving the files is on files

	// TODO: Save the files
	return responses, nil
}

// If there is no defference it will return nil too
func CompareTwoFiles() (*Response, error) {
	return nil, nil
}
