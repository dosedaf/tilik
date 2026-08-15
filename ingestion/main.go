package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"time"
	"strings"
)

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

	case "pemenang_berkontrak":
		if paths.PemenangBerkontrak == "" || kode == "" {
			return ""
		}

		path = fmt.Sprintf(paths.PemenangBerkontrak, kode)

	default:
		return ""
	}

	return fmt.Sprintf("%s/%s%s", baseURL, pemda, path)
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
		} else if limit == 0{
			ids = ids[:0] 
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

		c := newScraper(client, category)
		c2 := newScraper(client, category)

		var results []Paket
		var pemenangBerkontrak map[string]string
		var realisasi map[string][]Realisasi

		switch category {
		case "tender":
			results = scrapeTenderDetails(client, c, ids)
			pemenangBerkontrak = scrapeTenderPemenangBerkontrak(client, c2, ids)
		case "nontender":
			results = scrapeNonTenderDetails(client, c, ids)
			pemenangBerkontrak = scrapeNonTenderPemenangBerkontrak(client, c2, ids)
		case "pencatatan":
			results = scrapePencatatanDetails(client, c, ids)
			realisasi = scrapePencatatanRealisasi(client, c2, ids)
		case "swakelola":
			results = scrapeSwakelolaDetails(client, c, ids)
			realisasi = scrapeSwakelolaRealisasi(client, c2, ids)
		}

		if len(results) == 0 {
			printVerbose(
				"[%s] no package details scraped",
				category,
				)
			continue
		}

		switch category {
		case "tender":
			for i := range results {
				if results[i].Tender.PemenangBerkontrak == "Tender Batal" {
					continue
				}

				results[i].Tender.PemenangBerkontrak = pemenangBerkontrak[results[i].Kode]
			}
		case "nontender":
			for i := range results {
				results[i].NonTender.PemenangBerkontrak = pemenangBerkontrak[results[i].Kode]
			}
		case "pencatatan":
			for i := range results {
				results[i].Pencatatan.Realisasi =  realisasi[results[i].Kode]
			}
		case "swakelola":
			for i := range results {
				results[i].Swakelola.Realisasi = realisasi[results[i].Kode]
			}
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
