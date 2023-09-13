package main

import (
	"github.com/UTD-JLA/botsu/internal/data/vn"
)

func main() {
	titles, err := vn.ReadVNDBTitlesFile("vndb-db-2023-09-07/db/vn_titles")

	if err != nil {
		panic(err)
	}

	data, err := vn.ReadVNDBDataFile("vndb-db-2023-09-07/db/vn")

	if err != nil {
		panic(err)
	}

	vns := vn.JoinTitlesAndData(titles, data)

	searcher := vn.NewVNSearcher(vns)

	err = searcher.CreateIndex("vndb-index")

	if err != nil {
		panic(err)
	}

	_, err = searcher.LoadIndex("vndb-index")

	if err != nil {
		panic(err)
	}
}
