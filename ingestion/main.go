package main

import (
	"ingestion/internal/spse"
)

func main() {
	s := spse.NewSPSEScraper()

	pemdas := []string{
		"wonogirikab",
	}

	years := []string{
		"2025",
	}

	s.Scrape(pemdas, years)

}
