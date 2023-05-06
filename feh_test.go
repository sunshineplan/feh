package feh

import (
	"log"
	"testing"
)

func TestFEH(t *testing.T) {
	_, _, fullScoreboard, err := Scrape()
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range fullScoreboard {
		log.Print(i)
	}
}
