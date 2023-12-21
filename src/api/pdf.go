package api

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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

	// Use bash to execute the command with process substitution
	cmd := exec.Command("bash", "-c", fmt.Sprintf("diff %s %s", commands[0], commands[1]))

	output, err := cmd.CombinedOutput()
	if err != nil {
		if err.Error() == "exit status 1" {

			changes = parseDiff(string(output)).Changes

		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Println("No differences found.")
	}

	return &changes, nil
}

func parseDiff(diffOutput string) Changes {

	/*
		  This is a example of diff Output:
					19c19 -> this is the line number of the first file and then the second number is the line number of the second file
		  < another digit. If the dice is fair all six outcomes X = {1, . . . , 6} are equally likely to occur, hence we -> < represents what was
		  ---
		  > another digit. If the dice is fair all six outcomes X = {1, . . . . . , 6} are equally likely to occur, hence we -> > represents what is now
	*/

	lines := strings.Split(diffOutput, "\n")

	var changes Changes
	var currentChange Change

	for _, line := range lines {
		isChange, currentLines := getLineNumber(line)
		if isChange {
			for _, currentLine := range currentLines {
				lineNumber, _ := strconv.Atoi(currentLine)
				currentChange.Line = lineNumber
			}
		}
	}

	if currentChange.Change != "" {
		changes.Changes = append(changes.Changes, currentChange)
	}

	return changes
}

func isChangeLine(line string) bool {
	return strings.HasPrefix(line, "<") || strings.HasPrefix(line, ">") || regexp.MustCompile(`^\d+[a-c]\d+$`).MatchString(line)
}

func getLineNumber(line string) (bool, []string) {
	// This regex will get the number of the line. Example -> 19c19 it must get the two numbers, sometime is might be 19,19c19
	if regexp.MustCompile(`(?:^|[^<>])(\d+)`).MatchString(line) {
		return true, regexp.MustCompile(`(?:^|[^<>])(\d+)`).FindAllString(line, -1)
	}
	return false, nil
}
