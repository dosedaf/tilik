package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

var verbose = true

const (
	baseURL   = "https://spse.inaproc.id"
	portalID  = "slemankab"
	year      = "2026"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type DTResponse struct {
	Draw            interface{}     `json:"draw"`
	RecordsTotal    int             `json:"recordsTotal"`
	RecordsFiltered int             `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}

type TenderDetail struct {
	Kode              string
	NamaPaket         string
	Instansi          string
	SatuanKerja       string
	Kategori          string
	SistemPengadaan   string
	TahunAnggaran     string
	NilaiPagu         string
	NilaiHPS          string
	LokasiPekerjaan   string
	SyaratKualifikasi string
	URL               string
}

func printVerbose(format string, a ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", a...)
	}
}

func getSessionAndToken(client *http.Client) (string, error) {
	reqURL := fmt.Sprintf("%s/%s/lelang", baseURL, portalID)
	printVerbose("GET %s", reqURL)
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
	printVerbose("HTTP/1.1 %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))

	re := regexp.MustCompile(`authenticityToken\s*=\s*'([a-f0-9]+)'`)
	var bodyBuf strings.Builder
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyBuf.Write(buf[:n])
			matches := re.FindStringSubmatch(bodyBuf.String())
			if len(matches) > 1 {
				printVerbose("authenticityToken found: %s", matches[1])
				return matches[1], nil
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}

	return "", fmt.Errorf("authenticityToken not found")
}

func fetchTenderIDs(client *http.Client, token string) ([]string, error) {
	apiURL := fmt.Sprintf("%s/%s/dt/lelang?tahun=%s", baseURL, portalID, year)
	printVerbose("POST %s", apiURL)

	formData := url.Values{}
	formData.Set("draw", "1")
	formData.Set("start", "0")
	formData.Set("length", "-1")
	formData.Set("authenticityToken", token)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	printVerbose("HTTP/1.1 %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))

	var dtResp DTResponse
	if err := json.NewDecoder(resp.Body).Decode(&dtResp); err != nil {
		return nil, err
	}

	var tenderIDs []string
	for _, row := range dtResp.Data {
		if len(row) > 0 {
			switch val := row[0].(type) {
			case string:
				if val != "" {
					tenderIDs = append(tenderIDs, val)
				}
			case float64:
				tenderIDs = append(tenderIDs, fmt.Sprintf("%.0f", val))
			}
		}
	}

	printVerbose("payload extracted: %d records found.", len(tenderIDs))
	return tenderIDs, nil
}

func scrapeTenderDetails(client *http.Client, tenderIDs []string) []TenderDetail {
	var results []TenderDetail
	var mu sync.Mutex

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
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
		r.Headers.Set("Referer", fmt.Sprintf("%s/%s/lelang", baseURL, portalID))
		printVerbose("=> GET %s", r.URL.String())
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := TenderDetail{}
		detail.URL = e.Request.URL.String()

		// Extract Kode from URL fallback
		urlPath := e.Request.URL.Path
		re := regexp.MustCompile(`/lelang/(\d+)/pengumumanlelang`)
		matches := re.FindStringSubmatch(urlPath)
		if len(matches) > 1 {
			detail.Kode = matches[1]
		}

		wsRegex := regexp.MustCompile(`\s+`)
		
		// Fallback for Nama Paket if it's not in the table but in a header
		headerTitle := strings.TrimSpace(e.ChildText("h2, .page-header, strong"))
		if headerTitle != "" {
			detail.NamaPaket = wsRegex.ReplaceAllString(headerTitle, " ")
		}

		e.ForEach("table.table tr", func(_ int, row *colly.HTMLElement) {
			rawKey := strings.TrimSpace(row.ChildText("th"))
			rawVal := strings.TrimSpace(row.ChildText("td"))
			
			// Clean Whtiespace and Lowercase the key for robust matching
			keyLower := strings.ToLower(wsRegex.ReplaceAllString(rawKey, " "))
			val := wsRegex.ReplaceAllString(rawVal, " ")

			if keyLower == "" || val == "" {
				return
			}

			// Substring matching to catch inconsistencies like "nilai pagu paket "
			switch {
			case strings.Contains(keyLower, "kode tender"):
				detail.Kode = val
			case strings.Contains(keyLower, "nama paket"):
				detail.NamaPaket = val
			case strings.Contains(keyLower, "instansi"):
				detail.Instansi = val
			case strings.Contains(keyLower, "satuan kerja"):
				detail.SatuanKerja = val
			case strings.Contains(keyLower, "kategori"):
				detail.Kategori = val
			case strings.Contains(keyLower, "sistem pengadaan"):
				detail.SistemPengadaan = val
			case strings.Contains(keyLower, "tahun anggaran"):
				detail.TahunAnggaran = val
			case strings.Contains(keyLower, "nilai pagu paket") || strings.Contains(keyLower, "pagu"):
				detail.NilaiPagu = val
			case strings.Contains(keyLower, "nilai hps paket") || strings.Contains(keyLower, "hps"):
				detail.NilaiHPS = val
			case strings.Contains(keyLower, "lokasi pekerjaan") || strings.Contains(keyLower, "lokasi"):
				detail.LokasiPekerjaan = val
			}
		})

		var syarat []string
		e.ForEach("ul.syarat-kualifikasi li, div.kualifikasi-content", func(_ int, item *colly.HTMLElement) {
			text := strings.TrimSpace(item.Text)
			text = wsRegex.ReplaceAllString(text, " ")
			if text != "" {
				syarat = append(syarat, text)
			}
		})
		detail.SyaratKualifikasi = strings.Join(syarat, " | ")

		mu.Lock()
		results = append(results, detail)
		mu.Unlock()

		printVerbose("<= 200 OK %s", detail.Kode)
	})

	c.OnError(func(r *colly.Response, err error) {
		printVerbose("<= %d ERROR %s : %v", r.StatusCode, r.Request.URL.String(), err)
	})

	for _, id := range tenderIDs {
		targetURL := fmt.Sprintf("%s/%s/lelang/%s/pengumumanlelang", baseURL, portalID, id)
		c.Visit(targetURL)
	}

	c.Wait()
	return results
}

func exportToCSV(data []TenderDetail) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to export")
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("spse_%s_lelang_%s_%s.csv", portalID, year, timestamp)
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"Kode", "NamaPaket", "Instansi", "SatuanKerja", 
		"Kategori", "SistemPengadaan", "TahunAnggaran", 
		"NilaiPagu", "NilaiHPS", "LokasiPekerjaan", 
		"SyaratKualifikasi", "URL",
	}
	
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, d := range data {
		record := []string{
			d.Kode, d.NamaPaket, d.Instansi, d.SatuanKerja,
			d.Kategori, d.SistemPengadaan, d.TahunAnggaran,
			d.NilaiPagu, d.NilaiHPS, d.LokasiPekerjaan,
			d.SyaratKualifikasi, d.URL,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	printVerbose("Exported %d records to %s", len(data), filename)
	return nil
}

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		os.Exit(1)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	printVerbose("Resolving %s...", baseURL)
	token, err := getSessionAndToken(client)
	if err != nil {
		printVerbose("FATAL: %v", err)
		os.Exit(1)
	}

	tenderIDs, err := fetchTenderIDs(client, token)
	if err != nil {
		printVerbose("FATAL: %v", err)
		os.Exit(1)
	}

	if len(tenderIDs) == 0 {
		printVerbose("Exiting: 0 records found.")
		os.Exit(0)
	}

	results := scrapeTenderDetails(client, tenderIDs)

	err = exportToCSV(results)
	if err != nil {
		printVerbose("FATAL: %v", err)
		os.Exit(1)
	}
}
