package api

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	wg sync.WaitGroup
)

type Response struct {
	Id      string    `json:"id"`
	Author  string    `json:"author"`
	Files   []string  `json:"files"`
	Date    string    `json:"date"`
	Changes []Changes `json:"changes"`
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
func SaveFiles(files []*multipart.FileHeader) ([]*Response, error) {
	// Get the files from the API
	err := os.Mkdir("files", 0755)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	var responses []*Response
	for i, file := range files {
		var response Response

		if i%2 == 0 {
			response.Id = uuid.New().String()
			response.Files = append(response.Files, file.Filename)
			response.Date = time.Now().String()
			response.Changes = make([]Changes, 0)
			responses = append(responses, &response)
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

		_, err = os.Stat("files/" + response.Id + "/" + file.Filename)
		if !os.IsNotExist(err) {
			return nil, os.ErrExist
		}

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
		mainDir, err := os.Getwd()
		if err != nil {
			return err
		}
		mainDir = path.Join(mainDir, "files", r.Id)
		cmd := exec.Command("pdftotext", path.Join(mainDir, file.Name()), path.Join(mainDir, strings.Split(file.Name(), ".")[0]+".txt"))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	wg.Add(1)
	go r.deepCompare(files, &wg)
	wg.Wait()
	return nil
}

func (r *Response) deepCompare(files []fs.DirEntry, wg *sync.WaitGroup) {
	defer wg.Done()

	mainDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	txtFiles := []string{}
	for _, file := range files {
		txt := strings.Replace(file.Name(), ".pdf", ".txt", -1)
		txtFiles = append(txtFiles, txt)
	}

	mainDir = path.Join(mainDir, "files", r.Id)
	sf, err := os.Open(path.Join(mainDir, txtFiles[0]))
	if err != nil {
		log.Fatal(err)
	}
	defer sf.Close()

	df, err := os.Open(path.Join(mainDir, txtFiles[1]))
	if err != nil {
		log.Fatal(err)
	}
	defer df.Close()

	sscan := bufio.NewScanner(sf)
	dscan := bufio.NewScanner(df)

	lineNumber := 1
	for sscan.Scan() {
		dscan.Scan()
		if !bytes.Equal(sscan.Bytes(), dscan.Bytes()) {
			r.Changes = append(r.Changes, Changes{
				Original: Change{
					Line:    lineNumber,
					Content: sscan.Text(),
				},
				Difference: Change{
					Line:    lineNumber,
					Content: dscan.Text(),
				},
			})
		}
		lineNumber++
	}
	err = os.RemoveAll(mainDir)
	if err != nil {
		log.Fatal(err)
	}
}
