package anime

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type AniDBEntry struct {
	AID                   string
	PrimaryTitle          string
	XJatOfficialTitle     string
	XJatSynonyms          []string
	JapaneseOfficialTitle string
	JapaneseSynonyms      []string
	EnglishOfficialTitle  string
	EnglishSynonyms       []string
}

func ReadAniDBDump(path string) (entries []*AniDBEntry, err error) {
	entryMap := make(map[string]*AniDBEntry)

	file, err := os.Open(path)

	if err != nil {
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		if line[0] == '#' {
			continue
		}

		cols := strings.SplitN(line, "|", 4)

		if len(cols) != 4 {
			log.Println("[Warning] Invalid line:", line)
			continue
		}

		aid := cols[0]
		titleType := cols[1]
		language := cols[2]
		title := cols[3]

		if _, ok := entryMap[aid]; !ok {
			entryMap[aid] = &AniDBEntry{
				AID:              aid,
				XJatSynonyms:     []string{},
				JapaneseSynonyms: []string{},
				EnglishSynonyms:  []string{},
			}
		}

		entry := entryMap[aid]

		switch titleType {
		case "1":
			entry.PrimaryTitle = title
		case "4":
			switch language {
			case "x-jat":
				entry.XJatOfficialTitle = title
			case "ja":
				entry.JapaneseOfficialTitle = title
			case "en":
				entry.EnglishOfficialTitle = title
			}
		case "2":
			fallthrough
		case "3":
			switch language {
			case "x-jat":
				entry.XJatSynonyms = append(entry.XJatSynonyms, title)
			case "ja":
				entry.JapaneseSynonyms = append(entry.JapaneseSynonyms, title)
			case "en":
				entry.EnglishSynonyms = append(entry.EnglishSynonyms, title)
			}
		}

		entryMap[aid] = entry
	}

	if err = scanner.Err(); err != nil {
		return
	}

	for _, entry := range entryMap {
		entries = append(entries, entry)
	}

	return
}
