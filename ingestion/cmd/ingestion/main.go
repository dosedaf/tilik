package main

import (
	"ingestion/internal/spse"
)

func main() {
	s := spse.NewSPSEScraper()

// var pemdas = []string{
// 	// Banten
// 	"bantenprov",
// 	"pandeglangkab",
// 	"lebeskab",
// 	"serangkab",
// 	"tangerangkab",
// 	"lebakkab",
// 	"cilegon",
// 	"serangkota",
// 	"tangerangkota",
// 	"tangsel",
//
// 	// DKI Jakarta
// 	"jakarta",
//
// 	// Jawa Barat
// 	"jabarprov",
// 	"bogorkab",
// 	"sukabumikab",
// 	"cianjurkab",
// 	"bandungkab",
// 	"garutkab",
// 	"tasikmalayakab",
// 	"ciamis kab",
// 	"kuningankab",
// 	"cirebonkab",
// 	"majalengkakab",
// 	"sumedangkab",
// 	"indramayukab",
// 	"subang kab",
// 	"purwakartakab",
// 	"karawang kab",
// 	"bekasikab",
// 	"bandungbaratkab",
// 	"pangandarankab",
// 	"bogorkota",
// 	"sukabumikota",
// 	"bandungkota",
// 	"cimahikota",
// 	"tasikmalayakota",
// 	"banjarkota",
// 	"cirebonkota",
// 	"bekasikota",
// 	"depokkota",
//
// 	// Jawa Tengah
// 	"jatengprov",
// 	"cilacapkab",
// 	"banyumaskab",
// 	"purbalinggakab",
// 	"banjarnegarakab",
// 	"kebumenkab",
// 	"purworejokab",
// 	"wonosobokab",
// 	"magelangkab",
// 	"boyolalikab",
// 	"klatenkab",
// 	"sukoharjokab",
// 	"wonogirikab",
// 	"karanganyarkab",
// 	"sragenkab",
// 	"grobogankab",
// 	"blorakab",
// 	"rembangkab",
// 	"patikab",
// 	"ku skup",
// 	"demakkab",
// 	"semarangkab",
// 	"temanggungkab",
// 	"kendalkab",
// 	"batangkab",
// 	"pekalongankab",
// 	"pemalangkab",
// 	"tegalkab",
// 	"brebeskab",
// 	"magelangkota",
// 	"surakartakota",
// 	"salatiga",
// 	"semarangkota",
// 	"pekalongankota",
// 	"tegalkota",
//
// 	// DI Yogyakarta
// 	"yogyakarta",
// 	"slemankab",
// 	"bantulkab",
// 	"gunungkidulkab",
// 	"kulonprogokab",
//
// 	// Jawa Timur
// 	"jatimprov",
// 	"pacitankab",
// 	"ponorogokab",
// 	"trenggalekkab",
// 	"tulungagungkab",
// 	"blitarkab",
// 	"kedirikab",
// 	"malangkab",
// 	"lumajangkab",
// 	"jemberkab",
// 	"banyuwangikab",
// 	"bondowosokab",
// 	"situbondokab",
// 	"probolinggokab",
// 	"pasuruankab",
// 	"sidoarjokab",
// 	"mojokertokab",
// 	"jombangkab",
// 	"nganjukkab",
// 	"madiunkab",
// 	"magetankab",
// 	"ngawikab",
// 	"bojonegorokab",
// 	"tubankab",
// 	"lamongankab",
// 	"gresikkab",
// 	"bangkalankab",
// 	"sampangkab",
// 	"pamekasankab",
// 	"sumenepkab",
// 	"kediri kota",
// 	"blitarkota",
// 	"malangkota",
// 	"probolinggokota",
// 	"pasuruankota",
// 	"mojokertokota",
// 	"madiunkota",
// 	"surabayakota",
// 	"batukota",
//}

 pemdas := []string{"wonogirikab"}

	years := []string{
		"2025",
		// "2026",
	}

	s.Scrape(pemdas, years)
}
