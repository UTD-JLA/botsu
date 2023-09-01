package main

import (
	"log"

	"github.com/UTD-JLA/botsu/pkg/aodb"
)

func main() {
	err := aodb.ReadDatabaseFile("anime-offline-database.json")

	if err != nil {
		log.Fatal(err)
	}

	err = aodb.CreateIndex()

	if err != nil {
		log.Fatal(err)
	}
}
