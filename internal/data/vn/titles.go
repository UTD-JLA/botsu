package vn

import (
	"bufio"
	"os"
	"strings"
)

type VNTitle struct {
	Japanese string
	English  string
	Romaji   string
}

type titleLine struct {
	id       string
	language string
	official bool
	title    string
	latin    string
}

func ReadVNDBTitlesFile(path string) (map[string]*VNTitle, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	titles := make(map[string]*VNTitle)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if line[0] == '#' {
			continue
		}

		var title titleLine

		cols := strings.Split(line, "\t")

		title.id = cols[0]
		title.language = cols[1]
		title.official = cols[2] == "t"
		title.title = cols[3]

		if err != nil {
			return nil, err
		}

		if _, ok := titles[title.id]; !ok {
			titles[title.id] = &VNTitle{}
		}

		entry := titles[title.id]

		if title.language == "ja" {
			entry.Japanese = title.title

			if title.latin != "\\N" {
				entry.Romaji = title.latin
			}
		}

		if title.language == "en" && title.official {
			entry.English = title.title
		}
	}

	return titles, nil
}
