package main

import (
	"strings"
)

func pencatatanConfig() ScraperConfig {
	return ScraperConfig{
		Category:    "pencatatan",
		KodePrefix:  KodePrefix{
			Detail: "/pencatatan/",
		},
		InitDetail: func(url string) Paket {
			return Paket{Kategori: "pencatatan", URL: url, Pencatatan: &PencatatanDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "kode paket") },
				Handle: func(d *Paket, v string) { d.Kode = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama paket") },
				Handle: func(d *Paket, v string) { d.Nama = v },
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
				Handle: func(d *Paket, v string) { d.Pencatatan.JenisPengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "metode pengadaan") },
				Handle: func(d *Paket, v string) { d.Pencatatan.MetodePengadaan = v },
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
				},
			},
		},
	}
}


