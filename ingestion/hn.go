package main

import (
	"encoding/csv"
	"path/filepath"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
)

var verbose = true

const (
	baseURL   = "https://spse.inaproc.id"
	pemda     = "slemankab"
	year      = "2025"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

var scrapeLimits = map[string]int{
	"tender":     1,
	"nontender":  1,
	"swakelola":  1,
	"pencatatan": 1,
}

type CategoryPaths struct {
	Portal     string
	Pengumuman string
	Peserta    string
	Pemenang   string
	Dt         string
}

var categoryPaths = map[string]CategoryPaths{
	"tender": {
		Portal:     "/lelang",
		Pengumuman: "/lelang/%s/pengumumanlelang",
		Peserta:    "/lelang/%s/peserta",
		Pemenang:   "/evaluasilelang/%s/pemenang",
		Dt:         "/dt/lelang",
	},

	"nontender": {
		Portal:     "/nontender",
		Pengumuman: "/nontender/%s/pengumumanpl",
		Peserta:    "/nontender/%s/peserta",
		Pemenang:   "/evaluasinontender/%s/pemenang",
		Dt:         "/dt/pl",
	},

	"pencatatan": {
		Portal:     "/pencatatan",
		Pengumuman: "/pencatatan/pengumumannonspk?id=%s",
		Peserta:    "",
		Pemenang:   "/pencatatan/%s/pengumumannonspkpemenang",
		Dt:         "/dt/nonspk",
	},

	"swakelola": {
		Portal:     "/swakelola",
		Pengumuman: "/swakelola/%s/pengumuman",
		Peserta:    "",
		Pemenang:   "",
		Dt:         "/dt/swakelola",
	},
}

type DTResponse struct {
	Draw            interface{}     `json:"draw"`
	RecordsTotal    int             `json:"recordsTotal"`
	RecordsFiltered int             `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}

type PaketDetail struct {
	Kategori         string
	Kode             string
	NamaPaket        string
	KodeRUP          int
	SumberDana       string
	TanggalPembuatan string
	TahapTender      string
	Instansi         string
	SatuanKerja      string
	JenisPengadaan   string
	MetodePengadaan  string
	TahunAnggaran    string
	NilaiPagu        int64
	NilaiHPS         int64
	LokasiPekerjaan  string
	PesertaPaket     string
	URL              string
}

func printVerbose(format string, a ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", a...)
	}
}

func getPath(category, pemda, kode, page string) string {
	paths, ok := categoryPaths[category]
	if !ok {
		return ""
	}

	var path string

	switch page {
	case "portal":
		path = paths.Portal

	case "dt":
		path = paths.Dt

	case "pengumuman":
		if paths.Pengumuman == "" || kode == "" {
			return ""
		}

		path = fmt.Sprintf(paths.Pengumuman, kode)

	case "peserta":
		if paths.Peserta == "" || kode == "" {
			return ""
		}

		path = fmt.Sprintf(paths.Peserta, kode)

	case "pemenang":
		if paths.Pemenang == "" || kode == "" {
			return ""
		}

		path = fmt.Sprintf(paths.Pemenang, kode)

	default:
		return ""
	}

	return fmt.Sprintf("%s/%s%s", baseURL, pemda, path)
}

/*
func splitNumbers(s string) ([]int64, error) {
s = strings.TrimSpace(s)
s = strings.ReplaceAll(s, "Rp.", "")
s = strings.ReplaceAll(s, "Rp", "")

parts := strings.Fields(s)

var numbers []int64

for _, part := range parts {
part = strings.TrimSpace(part)

if part == "" {
continue
}

part = strings.ReplaceAll(part, ".", "")
part = strings.Split(part, ",")[0]

num, err := strconv.ParseInt(part, 10, 64)
if err != nil {
continue
}

numbers = append(numbers, num)
}

return numbers, nil
}
*/

func splitNumbers(s string) ([]int64, error) {
	re := regexp.MustCompile(`(?:Rp\.?\s*)?([\d.]+)(?:,\d+)?`)
	matches := re.FindAllStringSubmatch(s, -1)

	var numbers []int64

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		value := strings.ReplaceAll(match[1], ".", "")

		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}

		numbers = append(numbers, num)
	}

	return numbers, nil
}

func getSessionAndToken(client *http.Client, category string) (string, error) {
	reqURL := getPath(category, pemda, "", "portal")

	if reqURL == "" {
		return "", fmt.Errorf("invalid category: %s", category)
	}

	printVerbose("[%s] GET %s", category, reqURL)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	printVerbose(
		"[%s] HTTP/1.1 %d %s",
		category,
		resp.StatusCode,
		http.StatusText(resp.StatusCode),
		)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	//tokenRegex := regexp.MustCompile(`authenticityToken\s*=\s*'([a-f0-9]+)'`)
	tokenRegex := regexp.MustCompile(
		`authenticityToken\s*=\s*["']([a-fA-F0-9]+)["']`,
		)
	matches := tokenRegex.FindStringSubmatch(string(body))

	if len(matches) < 2 {
		return "", fmt.Errorf(
			"authenticityToken not found for category %s",
			category,
			)
	}

	printVerbose("[%s] authenticityToken found", category)

	return matches[1], nil
}

func fetchIDs(
	client *http.Client,
	token string,
	category string,
) ([]string, error) {
	apiURL := getPath(category, pemda, "", "dt")

	if apiURL == "" {
		return nil, fmt.Errorf("invalid DT path for category %s", category)
	}

	apiURL += "?tahun=" + url.QueryEscape(year)

	printVerbose("[%s] POST %s", category, apiURL)

	formData := url.Values{}

	formData.Set("draw", "1")
	formData.Set("start", "0")
	formData.Set("length", "-1")
	formData.Set("authenticityToken", token)

	req, err := http.NewRequest(
		"POST",
		apiURL,
		strings.NewReader(formData.Encode()),
		)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded; charset=UTF-8",
		)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"DT request failed: HTTP %d %s",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			)
	}

	printVerbose(
		"[%s] HTTP/1.1 %d %s",
		category,
		resp.StatusCode,
		http.StatusText(resp.StatusCode),
		)

	var dtResp DTResponse

	if err := json.NewDecoder(resp.Body).Decode(&dtResp); err != nil {
		return nil, err
	}

	var ids []string

	for _, row := range dtResp.Data {
		// printVerbose("[%s] ROW: %#v", category, row)
		if len(row) == 0 {
			continue
		}

		switch val := row[0].(type) {
		case string:
			if val != "" {
				ids = append(ids, val)
			}

		case float64:
			ids = append(ids, fmt.Sprintf("%.0f", val))
		}
	}

	printVerbose(
		"[%s] payload extracted: %d records found",
		category,
		len(ids),
		)

	return ids, nil
}

func scrapePaketDetails(
	client *http.Client,
	category string,
	ids []string,
) []PaketDetail {
	var results []PaketDetail
	var mu sync.Mutex
	var scraped atomic.Int64

	c := colly.NewCollector(
		colly.AllowedDomains("spse.inaproc.id"),
		colly.UserAgent(userAgent),
		colly.Async(true),
		)

	baseURLParsed, err := url.Parse(baseURL)

	if err == nil && client.Jar != nil {
		cookies := client.Jar.Cookies(baseURLParsed)

		if len(cookies) > 0 {
			_ = c.SetCookies(baseURL, cookies)
		}
	}

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*spse.inaproc.id*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set(
			"Accept",
			"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			)

		r.Headers.Set(
			"Accept-Language",
			"en-US,en;q=0.9,id;q=0.8",
			)

		r.Headers.Set(
			"Referer",
			getPath(category, pemda, "", "portal"),
			)

		printVerbose(
			"[%s] => GET %s",
			category,
			r.URL.String(),
			)
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := PaketDetail{
			Kategori: category,
			URL:      e.Request.URL.String(),
		}

		urlPath := e.Request.URL.Path

		switch {
		case strings.Contains(urlPath, "/pengumumanlelang"):
			detail.Kode = extractKode(urlPath, "/lelang/")

		case strings.Contains(urlPath, "/pengumumanpl"):
			detail.Kode = extractKode(urlPath, "/nontender/")

		case strings.Contains(urlPath, "/pengumumannonspk"):
			detail.Kode = e.Request.URL.Query().Get("id") 

		case strings.Contains(urlPath, "/swakelola/"):
			detail.Kode = extractKode(urlPath, "/swakelola/")
		}

		wsRegex := regexp.MustCompile(`\s+`)

		e.ForEach("table.table tr", func(_ int, row *colly.HTMLElement) {
			rawKey := strings.TrimSpace(
				row.ChildText("th"),
				)

			rawVal := strings.TrimSpace(
				row.ChildText("td"),
				)

			keyLower := strings.ToLower(
				wsRegex.ReplaceAllString(rawKey, " "),
				)

			val := wsRegex.ReplaceAllString(rawVal, " ")

			if keyLower == "" || val == "" {
				return
			}

			switch {
			case strings.EqualFold(keyLower, "kode tender"),
				strings.EqualFold(keyLower, "kode paket"),
				strings.EqualFold(keyLower, "kode swakelola"):
				detail.Kode = val

			case strings.EqualFold(keyLower, "nama tender"),
				strings.EqualFold(keyLower, "nama paket"),
				strings.EqualFold(keyLower, "nama swakelola"):
				detail.NamaPaket = val

			case strings.Contains(keyLower, "k/l/pd"):
				detail.Instansi = val

			case strings.Contains(keyLower, "satuan kerja"):
				detail.SatuanKerja = val

				if val == "1.02.0.00.0.00.01.0000" {
					detail.SatuanKerja = "Dinas Kesehatan"
				}

			case strings.Contains(keyLower, "jenis pengadaan"):
				detail.JenisPengadaan = val

			case strings.Contains(keyLower, "metode pengadaan"):
				detail.MetodePengadaan = val

			case strings.Contains(keyLower, "tahun anggaran"):
				detail.TahunAnggaran = val

			case strings.Contains(keyLower, "pagu"):
				numbers, err := splitNumbers(val)

				if err != nil {
					printVerbose(
						"[%s] failed to parse pagu: %v",
						category,
						err,
						)
					return
				}

				if len(numbers) >= 1 {
					detail.NilaiPagu = numbers[0]
				}

				if len(numbers) >= 2 {
					detail.NilaiHPS = numbers[1]
				}

			case strings.Contains(keyLower, "lokasi"):
				detail.LokasiPekerjaan = val
			}
		})

		if detail.Kode == "" {
			return
		}
		mu.Lock()
		results = append(results, detail)
		mu.Unlock()

		current := scraped.Add(1)

		printVerbose(
			"[%s] <= 200 OK %s %d/%d",
			category,
			detail.Kode,
			current,
			len(ids),
			)
	})

	c.OnError(func(r *colly.Response, err error) {
		printVerbose(
			"[%s] <= %d ERROR %s : %v",
			category,
			r.StatusCode,
			err,
			)
	})

	for _, id := range ids {
		targetURL := getPath(
			category,
			pemda,
			id,
			"pengumuman",
			)

		if targetURL == "" {
			printVerbose(
				"[%s] skipping ID %s: no pengumuman path",
				category,
				id,
				)
			continue
		}

		if err := c.Visit(targetURL); err != nil {
			printVerbose(
				"[%s] failed to visit %s: %v",
				category,
				targetURL,
				err,
				)
		}
	}

	c.Wait()

	return results
}

func exportToCSV(
	data []PaketDetail,
	category string,
) error {
	if len(data) == 0 {
		return fmt.Errorf(
			"no data to export for category %s",
			category,
			)
	}

	timestamp := time.Now().Format("20060102_150405")

	targetDir := "./spse"

	filename := fmt.Sprintf(
		"spse_%s_%s_%s.csv",
		pemda,
		category,
		timestamp,
		)

	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory: %v", err)
		return err
	}

	fullPath := filepath.Join(targetDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var headers []string

	if category == "tender" {
		headers = []string{
			"Kategori",
			"Kode Tender",
			"Nama Tender",
			"K/L/PD/Instansi Lainnya",
			"Satuan Kerja",
			"Jenis Pengadaan",
			"Metode Pengadaan",
			"Tahun Anggaran",
			"Nilai Pagu (dalam Rupiah)",
			"Nilai HPS (dalam Rupiah)",
			"Lokasi Pekerjaan",
			"URL",
		}
	} else if category == "nontender" {
		headers = []string{
			"Kategori",
			"Kode Paket",
			"Nama Paket",
			"K/L/PD/Instansi Lainnya",
			"Satuan Kerja",
			"Jenis Pengadaan",
			"Metode Pengadaan",
			"Tahun Anggaran",
			"Nilai Pagu (dalam Rupiah)",
			"Nilai HPS (dalam Rupiah)",
			"Lokasi Pekerjaan",
			"URL",
		}
	} else if category == "pencatatan" {
		headers = []string{
			"Kategori",
			"Kode Paket",
			"Nama Paket",
			"K/L/PD/Instansi Lainnya",
			"Satuan Kerja",
			"Jenis Pengadaan",
			"Metode Pengadaan",
			"Tahun Anggaran",
			"Nilai Pagu Paket (dalam Rupiah)",
			"URL",
		}
	} else if category == "swakelola" {
		headers = []string{
			"Kategori",
			"Kode Swakelola",
			"Nama Swakelola",
			"K/L/PD",
			"Satuan Kerja",
			// "Tipe Pelaksanaan Swakelola",
			"Tahun Anggaran",
			"Nilai Pagu Paket (dalam Rupiah)",
			"URL",
		}
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	var record []string

	for _, d := range data {
		if category == "tender" {
			record = []string{
				d.Kategori,
				d.Kode,
				d.NamaPaket,
				d.Instansi,
				d.SatuanKerja,
				d.JenisPengadaan,
				d.MetodePengadaan,
				d.TahunAnggaran,
				strconv.FormatInt(d.NilaiPagu, 10),
				strconv.FormatInt(d.NilaiHPS, 10),
				d.LokasiPekerjaan,
				d.URL,
			}
		} else if category == "nontender" {
			record = []string{
				d.Kategori,
				d.Kode,
				d.NamaPaket,
				d.Instansi,
				d.SatuanKerja,
				d.JenisPengadaan,
				d.MetodePengadaan,
				d.TahunAnggaran,
				strconv.FormatInt(d.NilaiPagu, 10),
				strconv.FormatInt(d.NilaiHPS, 10),
				d.LokasiPekerjaan,
				d.URL,
			}
		} else if category == "pencatatan" {
			record = []string{
				d.Kategori,
				d.Kode,
				d.NamaPaket,
				d.Instansi,
				d.SatuanKerja,
				d.JenisPengadaan,
				d.MetodePengadaan,
				d.TahunAnggaran,
				strconv.FormatInt(d.NilaiPagu, 10),
				d.URL,
			}
		} else if category == "swakelola" {
			record = []string{
				d.Kategori,
				d.Kode,
				d.NamaPaket,
				d.Instansi,
				d.SatuanKerja,
				d.TahunAnggaran,
				strconv.FormatInt(d.NilaiPagu, 10),
				d.URL,
			}
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	printVerbose(
		"[%s] Exported %d records to %s",
		category,
		len(data),
		filename,
		)

	return nil
}

func extractKode(path, prefix string) string {
	remaining := strings.TrimPrefix(path, prefix)

	if remaining == path {
		return ""
	}

	parts := strings.Split(remaining, "/")

	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

func main() {
	jar, err := cookiejar.New(nil)

	if err != nil {
		printVerbose("FATAL: %v", err)
		os.Exit(1)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	categories := []string{
		"tender",
		"nontender",
		"pencatatan",
		"swakelola",
	}

	for _, category := range categories {
		printVerbose("")
		printVerbose("==============================")
		printVerbose("CATEGORY: %s", category)
		printVerbose("==============================")

		token, err := getSessionAndToken(
			client,
			category,
			)

		if err != nil {
			printVerbose(
				"[%s] FATAL: %v",
				category,
				err,
				)
			continue
		}

		ids, err := fetchIDs(
			client,
			token,
			category,
			)

		if limit := scrapeLimits[category]; limit > 0 && len(ids) > limit {
			ids = ids[:limit]
		}

		if err != nil {
			printVerbose(
				"[%s] FATAL: %v",
				category,
				err,
				)
			continue
		}

		if len(ids) == 0 {
			printVerbose(
				"[%s] no records found",
				category,
				)
			continue
		}

		results := scrapePaketDetails(
			client,
			category,
			ids,
			)

		if len(results) == 0 {
			printVerbose(
				"[%s] no package details scraped",
				category,
				)
			continue
		}

		if err := exportToCSV(
			results,
			category,
			); err != nil {
			printVerbose(
				"[%s] export failed: %v",
				category,
				err,
				)
			continue
		}
	}

	printVerbose("")
	printVerbose("All categories completed.")
}
