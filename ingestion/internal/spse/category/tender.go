package category

import (
	"strings"
	"net/url"

	"ingestion/internal/spse/model"
	"ingestion/util"

	"github.com/gocolly/colly/v2"
)

func TenderConfig() ScraperConfig {
	kodePrefix := KodePrefix{
		Detail: "/lelang/",
		Evaluasi: "/evaluasi/",
	}
	return ScraperConfig{
		Category:    "tender",
		HasPemenang: true,
		HasPemenangBerkontrak: true,
		HasRealisasi: false,
		KodePrefix: kodePrefix,
		StatusIndex: 3,
		InitDetail: func(url string) model.Paket {
			return model.Paket{Kategori: "tender", URL: url, Tender: &model.TenderDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama tender") },
				Handle: func(d *model.Paket, v string) {
					// if strings.Contains(v, "Tender Batal") {
					// 	d.Tender.PemenangBerkontrak = "Tender Batal"
					// }
					// if strings.Contains(v, "Tender Gagal") {
					// 	d.Tender.PemenangBerkontrak = "Tender Gagal"
					// }
					// if strings.Contains(v, "Seleksi Batal") {
					// 	d.Tender.PemenangBerkontrak = "Seleksi Batal"
					// }
					// if strings.Contains(v, "Seleksi Gagal") {
					// 	d.Tender.PemenangBerkontrak = "Seleksi Gagal"
					// }
					// if strings.Contains(v, "Seleksi Ulang") {
					// 	d.Tender.PemenangBerkontrak = "Seleksi Ulang"
					// }
					// if strings.Contains(v, "Evaluasi Ulang") {
					// 	d.Tender.PemenangBerkontrak = "Evaluasi Ulang"
					// }
					// if strings.Contains(v, "Tender Ulang") {
					// 	d.Tender.PemenangBerkontrak = "Tender Ulang"
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
				Handle: func(d *model.Paket, v string) { d.Tender.JenisPengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "metode pengadaan") },
				Handle: func(d *model.Paket, v string) { d.Tender.MetodePengadaan = v },
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "reverse auction") },
				Handle: func(d *model.Paket, v string) { d.Tender.ReverseAuction = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "tahun anggaran") },
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
				Match:  func(k string) bool { return strings.EqualFold(k, "jenis kontrak") },
				Handle: func(d *model.Paket, v string) { d.Tender.JenisKontrak = v },
			},
			{
				Match:  func(k string) bool { return strings.Contains(k, "lokasi") },
				Handle: func(d *model.Paket, v string) { d.Tender.Lokasi = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "kualifikasi usaha") },
				Handle: func(d *model.Paket, v string) { d.Tender.KualifikasiUsaha = v },
			},
			{
				Match:  func(k string) bool { return strings.EqualFold(k, "peserta tender") },
				Handle: func(d *model.Paket, v string) {
					words := strings.Fields(v)
					d.Tender.Peserta = words[0]
				},
			},
			{
				Match: func(k string) bool {
					return strings.EqualFold(k, "rencana umum pengadaan")
				},

				HandleRow: func(d *model.Paket, row *colly.HTMLElement) {
					row.ForEach("table.table-sm > tbody > tr", func(i int, rupRow *colly.HTMLElement) {
						if i == 0 {
							return
						}

						cells := rupRow.DOM.ChildrenFiltered("td")

						if cells.Length() < 3 {
							return
						}

						kodeRUP := strings.TrimSpace(cells.Eq(0).Text())
						namaPaket := strings.TrimSpace(cells.Eq(1).Text())
						sumberDana := strings.TrimSpace(cells.Eq(2).Text())

						d.RUP.Kode = kodeRUP
						d.RUP.Nama = namaPaket
						d.RUP.SumberDana = sumberDana
					})
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
				if results[i].Tender.Pemenang != "" {
					continue
				}

				results[i].Tender.Pemenang = pemenang[results[i].Kode]

				if results[i].Tender.PemenangBerkontrak != "" {
					continue
				}

				results[i].Tender.PemenangBerkontrak = pemenangBerkontrak[results[i].Kode]
			}
		},
	}
}


