package category

import (
	"strings"

	"ingestion/internal/spse/model"
	"ingestion/util"
)

func PencatatanConfig() ScraperConfig {
	return ScraperConfig{
		Category:    "pencatatan",
		KodePrefix:  KodePrefix{
			Detail: "/pencatatan/",
		},
		InitDetail: func(url string) model.Paket {
			return model.Paket{Kategori: "pencatatan", URL: url, Pencatatan: &model.PencatatanDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "kode paket") },
				Handle: func(d *model.Paket, v string) { d.Kode = v },
			},
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama paket") },
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
				Match:  func(k string) bool { return strings.EqualFold(k, "jenis pengadaan") },
				Handle: func(d *model.Paket, v string) { d.Pencatatan.JenisPengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "metode pengadaan") },
				Handle: func(d *model.Paket, v string) { d.Pencatatan.MetodePengadaan = v },
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
	}
}


