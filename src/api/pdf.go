package api

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	wg sync.WaitGroup
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
	err := os.Mkdir("files", 0755)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	var responses []Response
	for i, file := range files {
		var response Response

		if i%2 == 0 {
			response.Id = uuid.New().String()
			response.Files = append(response.Files, file.Filename)
			response.Date = time.Now().String()
			responses = append(responses, response)
		} else {
			response.Id = responses[len(responses)-1].Id
			responses[len(responses)-1].Files = append(responses[len(responses)-1].Files, file.Filename)
		}

		err := os.Mkdir("files/"+response.Id, 0755)
		if err != nil && !os.IsExist(err) {
			return nil, err
		}

		// Save the file
		src, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		dst, err := os.Create("files/" + response.Id + "/" + file.Filename)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return nil, err
		}
		defer dst.Close()
	}

	// The Directory for saving the files is on files
	return responses, nil
}

// If there is no defference it will return nil too
func (r *Response) CompareTwoFilesInDir() error {

	// For each file in the folder of the ID we are creating a go routine that will compare the files
	files, err := os.ReadDir("files/" + r.Id)
	if err != nil {
		return err
	}

	for _, file := range files {
		// FIXME: Not pointing to the dir
		cmd := exec.Command("pdftotext " + file.Name() + file.Name() + ".txt")
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		os.Remove("files/" + r.Id + "/" + file.Name())
	}

	wg.Add(1)
	go r.deepCompare(files, &wg)
	wg.Wait()

	return nil
}

func (r *Response) deepCompare(files []fs.DirEntry, wg *sync.WaitGroup) bool {
	sf, err := os.Open("files/" + r.Id + "/" + files[0].Name())
	defer wg.Done()
	if err != nil {
		log.Fatal(err)
	}

	df, err := os.Open("files/" + r.Id + "/" + files[1].Name())
	if err != nil {
		log.Fatal(err)
	}

	sscan := bufio.NewScanner(sf)
	dscan := bufio.NewScanner(df)

	i := 0
	for sscan.Scan() {
		dscan.Scan()
		if !bytes.Equal(sscan.Bytes(), dscan.Bytes()) {
			fmt.Println(sscan.Text())
			fmt.Println(dscan.Text())
			r.Changes = append(r.Changes, Change{
				Line:    i,
				Content: sscan.Text(),
			})
			return true
		}
		i++
	}
	return false
}
