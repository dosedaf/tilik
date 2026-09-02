package category

import (
	"strings"
	"net/url"

	"ingestion/internal/spse/model"
	"ingestion/util"

	"github.com/gocolly/colly/v2"
)

func SwakelolaConfig() ScraperConfig {
	kodePrefix := KodePrefix{
			Detail: "/swakelola/",
			Evaluasi: "/pengumumanswakelolapelaksana/",
	}

	return ScraperConfig{
		Category:    "swakelola",
		HasPemenang: false,
		HasPemenangBerkontrak: false,
		HasRealisasi: true,
		KodePrefix:  kodePrefix,
		StatusIndex: 7,
		InitDetail: func(url string) model.Paket {
			return model.Paket{Kategori: "swakelola", URL: url, Swakelola: &model.SwakelolaDetail{}}
		},
		FieldRules: []FieldRule{
			{
				Match: func(k string) bool { return strings.EqualFold(k, "nama swakelola") },
				Handle: func(d *model.Paket, v string) { d.Nama = v },
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
				results[i].Swakelola.Realisasi = realisasi[results[i].Kode]
			}
		},
	}
}


