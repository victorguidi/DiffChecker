package api

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Changes struct {
	OriginalDoc string   `json:"originalDoc"`
	CompareDoc  string   `json:"compareDoc"`
	Changes     []Change `json:"changes"`
}

type Change struct {
	Change string
	Line   int
}

func CompareFiles(files ...string) (*[]Change, error) {

	defer func() {
		for _, file := range files {
			os.Remove(file)
		}
	}()

	commands := []string{}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println(err.Error())
			return nil, err
		}

		commands = append(commands, fmt.Sprintf("<(pdftotext -layout %s /dev/stdout)", file))
	}

	var changes []Change
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(len(commands))

	cmd := exec.Command("diff", commands[0], commands[1])
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	// Process the output as needed, for example, printing it
	fmt.Println(string(output))
	return &changes, nil
	for i, command := range commands {
		go func(command string, pos int, wg *sync.WaitGroup) {
			defer wg.Done()
			pos++
			if pos%2 != 0 {
				fmt.Println("Comparing files...")

				cmd := exec.Command("diff", commands[pos-1], commands[pos])
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					log.Fatal(err)
				}

				if err := cmd.Start(); err != nil {
					log.Fatal(err)
				}

				output, err := io.ReadAll(stdout)
				if err != nil {
					log.Fatal(err)
				}

				if err := cmd.Wait(); err != nil {
					log.Fatal(err)
				}

				// Process the output as needed, for example, printing it
				fmt.Println(string(output))

				// Modify 'changes' as needed
				mu.Lock()
				defer mu.Unlock()
				changes = append(changes, Change{Change: string(output), Line: 0})
			}
		}(command, i, &wg)
	}

	wg.Wait()

	return &changes, nil
}
