package vn

import (
	"bufio"
	"os"
	"strings"
)

const VNDBImageBaseURL = "https://t.vndb.org/cv/"

type VisualNovel struct {
	ID            string
	JapaneseTitle string
	EnglishTitle  string
	RomajiTitle   string
	Image         string
}

func (vn *VisualNovel) ImageURL() string {
	if vn.Image == "\\N" {
		return ""
	}

	imageID := vn.Image[2:]

	var lastTwoDigits string

	if len(imageID) > 1 {
		lastTwoDigits = imageID[len(imageID)-2:]
	} else {
		lastTwoDigits = "0" + imageID
	}

	return VNDBImageBaseURL + lastTwoDigits + "/" + imageID + ".jpg"
}

func JoinTitlesAndData(titles map[string]*VNTitle, data []*VNData) []*VisualNovel {
	vns := make([]*VisualNovel, 0, len(data))

	for _, entry := range data {
		vn := &VisualNovel{
			ID:    entry.ID,
			Image: entry.Image,
		}

		if title, ok := titles[entry.ID]; ok {
			vn.JapaneseTitle = title.Japanese
			vn.EnglishTitle = title.English
			vn.RomajiTitle = title.Romaji
		}

		vns = append(vns, vn)
	}

	return vns
}

// TODO: include more
type VNData struct {
	ID    string
	Image string
}

func ReadVNDBDataFile(path string) ([]*VNData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make([]*VNData, 0)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if line[0] == '#' {
			continue
		}

		var entry VNData

		cols := strings.Split(line, "\t")

		entry.ID = cols[0]
		entry.Image = cols[2]

		if err != nil {
			return nil, err
		}

		data = append(data, &entry)
	}

	return data, nil
}
