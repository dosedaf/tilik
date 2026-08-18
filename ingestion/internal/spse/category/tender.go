package category

import (
	"strings"
	"net/url"

	"ingestion/internal/spse/model"
	"ingestion/util"
)

func TenderConfig() ScraperConfig {
	kodePrefix := KodePrefix{
		Detail: "/lelang/",
		Evaluasi: "/evaluasi/",
	}
	return ScraperConfig{
		Category:    "tender",
		KodePrefix: kodePrefix,
		InitDetail: func(url string) model.Paket {
			return model.Paket{Kategori: "tender", URL: url, Tender: &model.TenderDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "kode tender") },
				Handle: func(d *model.Paket, v string) { d.Kode = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama tender") },
				Handle: func(d *model.Paket, v string) {
					if strings.Contains(v, "Tender Batal") {
						d.Tender.PemenangBerkontrak = "Tender Batal"
					}
					d.Nama = v
				},
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
				Match:  func(k string) bool { return strings.EqualFold(k, "jenis pengadaan") },
				Handle: func(d *model.Paket, v string) { d.Tender.JenisPengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "metode pengadaan") },
				Handle: func(d *model.Paket, v string) { d.Tender.MetodePengadaan = v },
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
					if len(numbers) >= 2 {
						d.Tender.HPS = numbers[1]
					}
				},
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "lokasi") },
				Handle: func(d *model.Paket, v string) { d.Tender.Lokasi = v },
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


