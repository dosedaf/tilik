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
