package main

import (
	"ingestion/internal/spse"
)

func main() {
	s := spse.NewSPSEScraper()

	pemdas := []string{
	//    "banjarnegarakab",
	//    "banyumaskab",
	//    "batangkab",
	//    "blorakab",
	//    "boyolali",
	//    "brebeskab",
	//    "cilacapkab",
	//    "demakkab",
	//    "grobogan",
	//    "jepara",
	"karanganyarkab",
	//    "kebumenkab",
	//    "kendalkab",
	// "klaten",
	 //    "kuduskab",
	// "magelangkab",
	//    "patikab",
	//    "pekalongankab",
	//    "pemalangkab",
	//    "purbalinggakab",
	//    "purworejokab",
	//    "rembangkab",
	//    "semarangkab",
	// "sragenkab",
	"sukoharjokab",
	//    "tegalkab",
	//    "temanggungkab",
	"wonogirikab",
	//    "wonosobokab",
	//
	//    "magelangkab",
	//    "pekalongankota",
	// "salatiga",
	//    "semarangkab",
	"surakarta",
	//    "tegalkab",
	}
	
	// pemda := []string{"wonogirikab"}

	years := []string{
		"2025",
		"2026",
	}

	s.Scrape(pemdas, years)
}
