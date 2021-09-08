package utils

import (
	"bufio"
	"fmt"
	"mydiary/models"
	"os"
	"strings"
)

func ReadFile(filename string) (*models.Content, error) {
	content := &models.Content{
		Filename: filename,
	}
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		// return nil, err
		return content, nil
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text = make(map[string]map[string]string)
	var currentDate string
	var currentBlock string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, " ")
		if strings.HasPrefix(line, "## ") {
			if currentDate != line[3:] {
				currentDate = line[3:]
				text[currentDate] = make(map[string]string)
				currentBlock = "Intro"
			}
			continue
		}
		if strings.HasPrefix(line, "#### ") {
			currentBlock = line[5:]
			continue
		}
		if currentBlock != "" {
			content.AppendLineToSection(currentDate, currentBlock, line)
		}
	}

	return content, nil
}

func WriteFile(content *models.Content) error {
	file, err := os.OpenFile(content.Filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	datawriter := bufio.NewWriter(file)

	defer closeAll(datawriter, file)

	for i, key := range content.Keys() {
		if i > 0 {
			datawriter.WriteString(fmt.Sprintln())
		}
		_, err := datawriter.WriteString(fmt.Sprintln("##", RemoveNewLine(key)))
		if err != nil {
			return err
		}
		sections, _ := content.GetSectionNames(key)
		for _, sectionName := range sections {
			if sectionName != "Intro" {
				_, err = datawriter.WriteString(fmt.Sprintln("####", RemoveNewLine(sectionName)))
				if err != nil {
					return err
				}
			}
			section, _ := content.GetSection(key, sectionName)
			_, err = datawriter.WriteString(fmt.Sprintln(RemoveNewLine(section.Value)))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RemoveNewLine(text string) string {
	return strings.TrimRight(text, "\n")
}

func closeAll(datawriter *bufio.Writer, file *os.File) {
	datawriter.Flush()
	file.Close()
}
