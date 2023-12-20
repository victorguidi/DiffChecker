package api

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

	// Use the sheel commands to compare the files
	// This is the shell command: diff <(pdftotext -layout t4.pdf /dev/stdout) <(pdftotext -layout t5.pdf /dev/stdout)
	// Extract the stdout from the shell command

	commands := []string{}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println(err.Error())
			return nil, err
		}

		commands = append(commands, fmt.Sprintf("<(pdftotext -layout %s /dev/stdout)", file))
	}

	// TODO: This should be a goroutine the command should be executed in parallel and the output should be appended to the result
	for _, command := range commands {
		go func(command string) {

			cmd := exec.Command("diff", commands[0], commands[1])
			cmd.Stdout = os.Stdout

			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}

		}(command)
	}

	return &[]Change{}, nil
}
