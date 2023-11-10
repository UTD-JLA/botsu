package vn

import (
	"bufio"
	"os"
	"strings"

	"github.com/blugelabs/bluge"
)

const VNDBImageBaseURL = "https://t.vndb.org/cv/"

const (
	SearchFieldJapaneseTitle = "japanese_title"
	SearchFieldEnglishTitle  = "english_title"
	SearchFieldRomajiTitle   = "romaji_title"
)

var SearchFields = []string{
	SearchFieldJapaneseTitle,
	SearchFieldEnglishTitle,
	SearchFieldRomajiTitle,
}

type VisualNovel struct {
	ID            string
	JapaneseTitle string
	EnglishTitle  string
	RomajiTitle   string
	Image         string
}

func (vn VisualNovel) Marshal() (*bluge.Document, error) {
	doc := bluge.NewDocument(vn.ID)

	doc.AddField(bluge.NewTextField(SearchFieldJapaneseTitle, vn.JapaneseTitle).StoreValue())
	doc.AddField(bluge.NewTextField(SearchFieldEnglishTitle, vn.EnglishTitle).StoreValue())
	doc.AddField(bluge.NewTextField(SearchFieldRomajiTitle, vn.RomajiTitle).StoreValue())

	return doc, nil
}

func (vn *VisualNovel) Unmarshal(fields map[string][]byte) error {
	vn.ID = string(fields["_id"])
	vn.JapaneseTitle = string(fields[SearchFieldJapaneseTitle])
	vn.EnglishTitle = string(fields[SearchFieldEnglishTitle])
	vn.RomajiTitle = string(fields[SearchFieldRomajiTitle])

	return nil
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
