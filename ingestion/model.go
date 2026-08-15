package main

import (
	"time"
)

var verbose = true

const (
	baseURL   = "https://spse.inaproc.id"
	pemda     = "slemankab"
	year      = "2025"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

var scrapeLimits = map[string]int{
	"tender":     0,
	"nontender":  0,
	"pencatatan": 50,
	"swakelola":  50,
}

type CategoryPaths struct {
	Portal     string
	Pengumuman string
	Peserta    string
	Pemenang   string
	PemenangBerkontrak string
	Dt         string
}

var categoryPaths = map[string]CategoryPaths{
	"tender": {
		Portal:     "/lelang",
		Pengumuman: "/lelang/%s/pengumumanlelang",
		Peserta:    "/lelang/%s/peserta",
		Pemenang:   "/evaluasi/%s/pemenang",
		PemenangBerkontrak:   "/evaluasi/%s/pemenangberkontrak",
		Dt:         "/dt/lelang",
	},

	"nontender": {
		Portal:     "/nontender",
		Pengumuman: "/nontender/%s/pengumumanpl",
		Peserta:    "/nontender/%s/peserta",
		Pemenang:   "/evaluasinontender/%s/pemenang",
		PemenangBerkontrak:   "/evaluasinontender/%s/pemenangberkontrak",
		Dt:         "/dt/pl",
	},

	"pencatatan": {
		Portal:     "/pencatatan",
		Pengumuman: "/pencatatan/pengumumannonspk?id=%s",
		// https://spse.inaproc.id/slemankab/pencatatan/pengumumannonspkpemenang?id=10186219000
		PemenangBerkontrak:   "/pencatatan/pengumumannonspkpemenang?id=%s",
		Dt:         "/dt/nonspk",
	},

	"swakelola": {
		Portal:     "/swakelola",
		Pengumuman: "/swakelola/%s/pengumuman",
		Peserta:    "",
		PemenangBerkontrak:   "/swakelola/pengumumanswakelolapelaksana/%s",
		Dt:         "/dt/swakelola",
	},
}

type DTResponse struct {
	Draw            interface{}     `json:"draw"`
	RecordsTotal    int             `json:"recordsTotal"`
	RecordsFiltered int             `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}


type Paket struct {
	Kode	string
	Kategori string
	Tahun	string
	TanggalPembuatan time.Time
	Nama string
	Instansi string
	KodeRUP string
	Satker string
	Pagu int64
	SumberDana string
	URL string

	Tender *TenderDetail
	NonTender *NonTenderDetail
	Pencatatan *PencatatanDetail
	Swakelola *SwakelolaDetail
}

type TenderDetail struct {
	JenisPengadaan	string
	MetodePengadaan string
	//JenisKontrak string
	HPS int64
	Lokasi string
	PemenangBerkontrak string
}

type NonTenderDetail struct {
	JenisPengadaan	string
	MetodePengadaan string
	JenisKontrak string
	HPS int64
	Lokasi string
	PemenangBerkontrak string
}

type PencatatanDetail struct {
	JenisPengadaan	string
	MetodePengadaan string
	PemenangBerkontrak []string
	Realisasi []Realisasi
}

type SwakelolaDetail struct {
	TipePelaksana string
	Pelaksana []Realisasi
	Realisasi []Realisasi
}

type Realisasi struct{
	Jenis string
	Nilai string
	Tanggal time.Time
}
