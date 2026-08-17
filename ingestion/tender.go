package main

import (
	"strings"
)

func tenderConfig() ScraperConfig {
	return ScraperConfig{
		Category:    "tender",
		KodePrefix:  KodePrefix{
			Detail: "/lelang/",
			Evaluasi: "/evaluasi/",
		},
		InitDetail: func(url string) Paket {
			return Paket{Kategori: "tender", URL: url, Tender: &TenderDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "kode tender") },
				Handle: func(d *Paket, v string) { d.Kode = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama tender") },
				Handle: func(d *Paket, v string) {
					if strings.Contains(v, "Tender Batal") {
						d.Tender.PemenangBerkontrak = "Tender Batal"
					}
					d.Nama = v
				},
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "k/l/pd") },
				Handle: func(d *Paket, v string) { d.Instansi = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "satuan kerja") },
				Handle: func(d *Paket, v string) {
					d.Satker = v
					if v == "1.02.0.00.0.00.01.0000" {
						d.Satker = "Dinas Kesehatan"
					}
				},
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "jenis pengadaan") },
				Handle: func(d *Paket, v string) { d.Tender.JenisPengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "metode pengadaan") },
				Handle: func(d *Paket, v string) { d.Tender.MetodePengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "tahun anggaran") },
				Handle: func(d *Paket, v string) { d.Tahun = v },
			},
			{
				Match: func(k string) bool { return strings.Contains(k, "pagu") },
				Handle: func(d *Paket, v string) {
					numbers, err := splitNumbers(v)
					if err != nil {
						printVerbose("[tender] failed to parse pagu: %v", err)
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
				Handle: func(d *Paket, v string) { d.Tender.Lokasi = v },
			},
		},
	}
}


