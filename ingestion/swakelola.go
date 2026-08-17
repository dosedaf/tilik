package main

import (
	"strings"
)

func swakelolaConfig() ScraperConfig {
	return ScraperConfig{
		Category:    "swakelola",
		KodePrefix:  KodePrefix{
			Detail: "/swakelola/",
			Evaluasi: "pengumumanswakelolapelaksana",
		},
		InitDetail: func(url string) Paket {
			return Paket{Kategori: "swakelola", URL: url, Swakelola: &SwakelolaDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "kode swakelola") },
				Handle: func(d *Paket, v string) { d.Kode = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama swakelola") },
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
				Match: func(k string) bool { return strings.EqualFold(k, "tipe pelaksana swakelola") },
				Handle: func(d *Paket, v string) { d.Tahun = v },
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


