package spse

import (
	"encoding/csv"
	"path/filepath"
	"fmt"
	"strings"
	"os"
	"strconv"

	"ingestion/internal/spse/model"
	"ingestion/util"
)

func (s *SPSEScraper) ExportToCSV(ctx ScrapeContext, data []model.Paket, category string) error {

	if len(data) == 0 {
		return fmt.Errorf(
			"no data to export for category %s",
			category,
			)
	}

	targetDir := filepath.Join(
		"../data/spse",
		ctx.Pemda,
		ctx.Year,
		ctx.IngestionID,
		)

	filename := category + ".csv"
	path := filepath.Join(targetDir, filename)

	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory: %v", err)
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var headers []string

	switch category {
	case "tender":
		headers = []string{
			"Kategori",
			"Kode Tender",
			"Nama Tender",
			"Kode RUP",
			"Sumber Dana",
			"Tanggal Pembuatan",
			"K/L/PD/Instansi Lainnya",
			"Status",
			"Satuan Kerja",
			"Jenis Pengadaan",
			"Metode Pengadaan",
			"Reverse Auction",
			"Tahun Anggaran",
			"Nilai Pagu (dalam Rupiah)",
			"Nilai HPS (dalam Rupiah)",
			"Jenis Kontrak",
			"Lokasi Pekerjaan",
			"Kualifikasi Usaha",
			"Jumlah Peserta",
			"Pemenang",
			"Pemenang Berkontrak",
			"URL",
		}
	case "nontender":
		headers = []string{
			"Kategori",
			"Kode Paket",
			"Nama Paket",
			"Tanggal Pembuatan",
			"K/L/PD/Instansi Lainnya",
			"Status",
			"Satuan Kerja",
			"Jenis Pengadaan",
			"Metode Pengadaan",
			"Khusus Orang Asli Papua (OAP)",
			"Tahun Anggaran",
			"Nilai Pagu (dalam Rupiah)",
			"Nilai HPS (dalam Rupiah)",
			"Jenis Kontrak",
			"Lokasi Pekerjaan",
			"Jumlah Peserta",
			"Pemenang",
			"Pemenang Berkontrak",
			"URL",
		}
	case "pencatatan":
		headers = []string{
			"Kategori",
			"Kode Paket",
			"Nama Paket",
			"Tanggal Pembuatan",
			"K/L/PD/Instansi Lainnya",
			"Status",
			"Satuan Kerja",
			"Jenis Pengadaan",
			"Metode Pengadaan",
			"Tahun Anggaran",
			"Nilai Pagu Paket (dalam Rupiah)",
			"Realisasi",
			"URL",
		}
	case "swakelola":
		headers = []string{
			"Kategori",
			"Kode Swakelola",
			"Nama Swakelola",
			"Tanggal Pembuatan",
			"K/L/PD",
			"Status",
			"Satuan Kerja",
			// "Tipe Pelaksanaan Swakelola",
			"Tahun Anggaran",
			"Nilai Pagu Paket (dalam Rupiah)",
			"Realisasi",
			"URL",
		}
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	var record []string

	for _, d := range data {
		switch category {
		case "tender":
			record = []string{
				d.Kategori,
				d.Kode,
				d.Nama,
				d.RUP.Kode,
				d.RUP.SumberDana,
				d.TanggalPembuatan,
				d.Instansi,
				d.Status,
				d.Satker,
				d.Tender.JenisPengadaan,
				d.Tender.MetodePengadaan,
				d.Tender.ReverseAuction,
				d.Tahun,
				strconv.FormatInt(d.Pagu, 10),
				strconv.FormatInt(d.Tender.HPS, 10),
				d.Tender.JenisKontrak,
				d.Tender.Lokasi,
				d.Tender.KualifikasiUsaha,
				d.Tender.Peserta,
				d.Tender.Pemenang,
				d.Tender.PemenangBerkontrak,
				d.URL,
			}
		case "nontender":
			record = []string{
				d.Kategori,
				d.Kode,
				d.Nama,
				d.TanggalPembuatan,
				d.Instansi,
				d.Status,
				d.Satker,
				d.NonTender.JenisPengadaan,
				d.NonTender.MetodePengadaan,
				d.NonTender.OAP,
				d.Tahun,
				strconv.FormatInt(d.Pagu, 10),
				strconv.FormatInt(d.NonTender.HPS, 10),
				d.NonTender.JenisKontrak,
				d.NonTender.Lokasi,
				d.NonTender.Peserta,
				d.NonTender.PemenangBerkontrak,
				d.NonTender.PemenangBerkontrak,
				d.URL,
			}
		case "pencatatan":
			var strSlice []string
			strSlice = []string{}
			realisasi := ""
			ada := false
			num := 0

			if len(d.Pencatatan.Realisasi) == 0 {
				ada = false
			} else {
				ada = true
				for i := range d.Pencatatan.Realisasi {
					num++
					p := d.Pencatatan.Realisasi[i]
					str := fmt.Sprintf("%d. Jenis: %s\nNilai: %s\nTanggal: %s",
						num,
						p.Jenis,
						p.Nilai,
						p.Tanggal,
						)

					strSlice = append(strSlice, str)
				}
			}

				if ada {
					realisasi = strings.Join(strSlice, "\n")
				} else {
					realisasi = ""
				}

				record = []string{
					d.Kategori,
					d.Kode,
					d.Nama,
					d.TanggalPembuatan,
					d.Instansi,
					d.Status,
					d.Satker,
					d.Pencatatan.JenisPengadaan,
					d.Pencatatan.MetodePengadaan,
					d.Tahun,
					strconv.FormatInt(d.Pagu, 10),
					realisasi,
					d.URL,
				}
		case "swakelola":
			var strSlice []string
			strSlice = []string{}
			realisasi := ""
			ada := false
			num := 0

			if len(d.Swakelola.Realisasi) == 0 {
				ada = false
			} else {
				ada = true
				for i := range d.Swakelola.Realisasi {
					num++
					p := d.Swakelola.Realisasi[i]
					str := fmt.Sprintf("%d. Jenis: %s\nNilai: %s\nTanggal: %s",
						num,
						p.Jenis,
						p.Nilai,
						p.Tanggal,
						)

					strSlice = append(strSlice, str)
				}
			}

				if ada {
					realisasi = strings.Join(strSlice, "\n")
				} else {
					realisasi = ""
				}

			record = []string{
				d.Kategori,
				d.Kode,
				d.Nama,
				d.TanggalPembuatan,
				d.Instansi,
				d.Status,
				d.Satker,
				d.Tahun,
				strconv.FormatInt(d.Pagu, 10),
				realisasi,
				d.URL,
			}

		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	util.PrintVerbose(
		"[%s] Exported %d records to %s",
		category,
		len(data),
		filename,
		)

	return nil
}
