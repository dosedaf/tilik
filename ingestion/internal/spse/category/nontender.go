package category

import (
	"strings"
	"net/url"

	"ingestion/internal/spse/model"
	"ingestion/util"
)

func NonTenderConfig() ScraperConfig {
	kodePrefix := KodePrefix{
			Detail: "/nontender/",
			Evaluasi: "/evaluasinontender/",
	}

	return ScraperConfig{
		Category:    "nontender",
		HasPemenang: true,
		HasPemenangBerkontrak: true,
		HasRealisasi: false,
		KodePrefix:  kodePrefix,
		StatusIndex: 3,
		InitDetail: func(url string) model.Paket {
			return model.Paket{Kategori: "nontender", URL: url, NonTender: &model.NonTenderDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama paket") },
				Handle: func(d *model.Paket, v string) { 
					// if strings.Contains(v, "Paket Gagal") {
					// 	d.NonTender.PemenangBerkontrak = "Paket Gagal"
					// }
					//
					// if strings.Contains(v, "Paket Batal") {
					// 	d.NonTender.PemenangBerkontrak = "Paket Batal"
					// }
					//
					// if strings.Contains(v, "Paket Ulang") {
					// 	d.NonTender.PemenangBerkontrak = "Paket Ulang"
					// }
					d.Nama = v 
				},
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "tanggal pembuatan") },
				Handle: func(d *model.Paket, v string) { d.TanggalPembuatan = v },
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
				Handle: func(d *model.Paket, v string) { d.NonTender.JenisPengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "metode pengadaan") },
				Handle: func(d *model.Paket, v string) { d.NonTender.MetodePengadaan = v },
			},
			{ // sementara OAP string dulu, nanti kalo fix Ya Tidak ganti ke bool dgn if block
				Match:  func(k string) bool { return strings.Contains(k, "khusus orang asli papua") },
				Handle: func(d *model.Paket, v string) { 
					d.NonTender.OAP = v
				},
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
						d.NonTender.HPS = numbers[1]
					}
				},
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "jenis kontrak") },
				Handle: func(d *model.Paket, v string) { d.NonTender.JenisKontrak = v },
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "lokasi") },
				Handle: func(d *model.Paket, v string) { d.NonTender.Lokasi = v },
			},
{
				Match:  func(k string) bool { return strings.EqualFold(k, "peserta tender") },
				Handle: func(d *model.Paket, v string) {
					words := strings.Fields(v)
					d.Tender.Peserta = words[0]
				},
			},

		},
		ExtractDetailKode: func(u *url.URL) string {
			return(extractKode(u, kodePrefix.Detail))
		},
		ExtractEvaluasiKode: func(u *url.URL) string {
			return(extractKode(u, kodePrefix.Evaluasi))
		},
		Enrich: func(results []model.Paket, pemenang map[string]string, pemenangBerkontrak map[string]string, realisasi map[string][]model.Realisasi) {
			for i := range results {
				if results[i].NonTender.Pemenang != "" {
					continue
				}

				results[i].NonTender.Pemenang = pemenang[results[i].Kode]

				if results[i].NonTender.PemenangBerkontrak != "" {
					continue
				}

				results[i].NonTender.PemenangBerkontrak = pemenangBerkontrak[results[i].Kode]
			}
		},
	}
}


