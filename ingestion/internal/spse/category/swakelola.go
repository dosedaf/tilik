package category

import (
	"strings"
	"net/url"

	"ingestion/internal/spse/model"
	"ingestion/util"
)

func SwakelolaConfig() ScraperConfig {
	kodePrefix := KodePrefix{
			Detail: "/swakelola/",
			Evaluasi: "/pengumumanswakelolapelaksana/",
	}

	return ScraperConfig{
		Category:    "swakelola",
		KodePrefix:  kodePrefix,
		InitDetail: func(url string) model.Paket {
			return model.Paket{Kategori: "swakelola", URL: url, Swakelola: &model.SwakelolaDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "kode swakelola") },
				Handle: func(d *model.Paket, v string) { d.Kode = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama swakelola") },
				Handle: func(d *model.Paket, v string) { d.Nama = v },
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "k/l/pd") },
				Handle: func(d *model.Paket, v string) { d.Instansi = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "satuan kerja") },
				Handle: func(d *model.Paket, v string) {
					d.Satker = v
					if v == "1.02.0.00.0.00.01.0000" {
						d.Satker = "Dinas Kesehatan"
					}
				},
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "tipe pelaksana swakelola") },
				Handle: func(d *model.Paket, v string) { d.Tahun = v },
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "tahun anggaran") },
				Handle: func(d *model.Paket, v string) { d.Tahun = v },
			},
			{
				Match: func(k string) bool { return strings.Contains(k, "pagu") },
				Handle: func(d *model.Paket, v string) {
					numbers, err := util.SplitNumbers(v)
					if err != nil {
						util.PrintVerbose("[tender] failed to parse pagu: %v", err)
						return
					}
					if len(numbers) >= 1 {
						d.Pagu = numbers[0]
					}
				},
			},
		},
		ExtractDetailKode: func(u *url.URL) string {
			return(extractKode(u, kodePrefix.Detail))
		},
		ExtractEvaluasiKode: func(u *url.URL) string {
			return(extractKode(u, kodePrefix.Evaluasi))
		},
	}
}


