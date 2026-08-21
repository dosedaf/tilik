package category

import (
	"net/url"
	"ingestion/internal/spse/model"
)

type FieldRule struct {
	Match  func(keyLower string) bool
	Handle func(detail *model.Paket, val string)
}

type ScraperConfig struct {
	Category    string
	KodePrefix KodePrefix
	InitDetail  func(reqURL string) model.Paket 
	FieldRules  []FieldRule

	ExtractDetailKode func(*url.URL) string
	ExtractEvaluasiKode func(*url.URL) string

	HasPemenang bool
	HasPemenangBerkontrak bool
	HasRealisasi bool

	Enrich func(
		results []model.Paket,
		pemenang map[string]string,
		pemenangBerkontrak map[string]string,
		realisasi map[string][]model.Realisasi,
	)
}

type KodePrefix struct {
	Detail  string 
	Evaluasi string
}

func AllConfigs() []ScraperConfig {
	return []ScraperConfig{
		TenderConfig(),
		NonTenderConfig(),
		PencatatanConfig(),
		SwakelolaConfig(),
	}
}
