package main 

import (
	"fmt"

	"github.com/gocolly/colly/v2"
)

var BaseURL string = "https://spse.inaproc.id"
var Pemda string =  "slemankab"

var Tipe = []string{"lelang", "nontender", "pencatatan", "swakelola", "darurat"}

// https://spse.inaproc.id/slemankab/nontender/10059063000/pengumumanpl

type PaketLPSE struct {
	KodePaket             string                `json:"kode_paket"`
	NamaPaket             string                `json:"nama_paket"`
	RUP                   []RUP                `json:"rup"`
	UraianSingkatPekerjaan string               `json:"uraian_singkat_pekerjaan"`
	TanggalPembuatan      string                `json:"tanggal_pembuatan"`
	TahapPaket            string                `json:"tahap_paket"`
	Instansi              string                `json:"instansi"`
	SatuanKerja           string                `json:"satuan_kerja"`
	JenisPengadaan        string                `json:"jenis_pengadaan"`
	MetodePengadaan       string                `json:"metode_pengadaan"`
	KhususOAP             bool                  `json:"khusus_oap"`
	TahunAnggaran         string                `json:"tahun_anggaran"`
	NilaiPagu             int64                 `json:"nilai_pagu"`
	NilaiHPS              int64                 `json:"nilai_hps"`
	JenisKontrak          string                `json:"jenis_kontrak"`
	Lokasi                []Lokasi              `json:"lokasi"`
	Kualifikasi           Kualifikasi           `json:"kualifikasi"`
	Peserta               []PesertaNonTender    `json:"peserta"`
	Pemenang              Pemenang             `json:"pemenang"`
}

type RUP struct {
	KodeRUP    string `json:"kode_rup"`
	NamaPaket  string `json:"nama_paket"`
	SumberDana string `json:"sumber_dana"`
}

type Lokasi struct {
	Kabupaten string `json:"kabupaten"`
}

type Kualifikasi struct {
	IzinUsaha      []IzinUsaha `json:"izin_usaha"`
	Administrasi   []string    `json:"administrasi"`
	Teknis         []string    `json:"teknis"`
}

type IzinUsaha struct {
	Jenis  string `json:"jenis"`
	Bidang string `json:"bidang"`
}

type PesertaNonTender struct {
	Nama             string `json:"nama"`
	NPWP             string `json:"npwp"`
	HargaPenawaran   int64  `json:"harga_penawaran"`
	HargaTerkoreksi  int64  `json:"harga_terkoreksi"`
	HargaNegosiasi   int64  `json:"harga_negosiasi"`
}

type Pemenang struct {
	Nama             string `json:"nama"`
	Alamat           string `json:"alamat"`
	NPWP             string `json:"npwp"`
	HargaPenawaran   int64  `json:"harga_penawaran"`
	HargaTerkoreksi  int64  `json:"harga_terkoreksi"`
	HasilNegosiasi   int64  `json:"hasil_negosiasi"`
}


func c() {
	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("http://go-colly.org/")
}
