package main

import (
	"ingestion/internal/spse"
)

func main() {
	s := spse.NewSPSEScraper()

	pemdas := []string{
		"wonogirikab",
	}
	s.Scrape(pemdas)

}
