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

	"ingestion/internal/spse"
	"ingestion/internal/spse/model"
	"ingestion/util"
)

func getSessionAndToken(client *http.Client, category string) (string, error) {
	reqURL := spse.GetPath(category, model.Pemda, "", "portal")

	if reqURL == "" {
		return "", fmt.Errorf("invalid category: %s", category)
	}

	util.PrintVerbose("[%s] GET %s", category, reqURL)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", model.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	util.PrintVerbose(
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

	util.PrintVerbose("[%s] authenticityToken found", category)

	return matches[1], nil
}

func fetchIDs(
	client *http.Client,
	token string,
	category string,
) ([]string, error) {
	apiURL := spse.GetPath(category, model.Pemda, "", "dt")

	if apiURL == "" {
		return nil, fmt.Errorf("invalid DT path for category %s", category)
	}

	apiURL += "?tahun=" + url.QueryEscape(model.Year)

	util.PrintVerbose("[%s] POST %s", category, apiURL)

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
	req.Header.Set("User-Agent", model.UserAgent)

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

	util.PrintVerbose(
		"[%s] HTTP/1.1 %d %s",
		category,
		resp.StatusCode,
		http.StatusText(resp.StatusCode),
		)

	var dtResp model.DTResponse

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

	util.PrintVerbose(
		"[%s] payload extracted: %d records found",
		category,
		len(ids),
		)

	return ids, nil
}

func main() {
	jar, err := cookiejar.New(nil)

	if err != nil {
		util.PrintVerbose("FATAL: %v", err)
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
		util.PrintVerbose("")
		util.PrintVerbose("==============================")
		util.PrintVerbose("CATEGORY: %s", category)
		util.PrintVerbose("==============================")

		token, err := getSessionAndToken(
			client,
			category,
			)

		if err != nil {
			util.PrintVerbose(
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

		if limit := model.ScrapeLimits[category]; limit > 0 && len(ids) > limit {
			ids = ids[:limit]
		} else if limit == 0{
			ids = ids[:0] 
		}

		if err != nil {
			util.PrintVerbose(
				"[%s] FATAL: %v",
				category,
				err,
				)
			continue
		}

		if len(ids) == 0 {
			util.PrintVerbose(
				"[%s] no records found",
				category,
				)
			continue
		}

		c := spse.NewScraper(client, category)
		c2 := spse.NewScraper(client, category)

		var results []model.Paket
		var pemenangBerkontrak map[string]string
		var realisasi map[string][]model.Realisasi

		switch category {
		case "tender":
			results = spse.ScrapeTenderDetails(client, c, ids)
			pemenangBerkontrak = spse.ScrapeTenderPemenangBerkontrak(client, c2, ids)
		case "nontender":
			results = spse.ScrapeNonTenderDetails(client, c, ids)
			pemenangBerkontrak = spse.ScrapeNonTenderPemenangBerkontrak(client, c2, ids)
		case "pencatatan":
			results = spse.ScrapePencatatanDetails(client, c, ids)
			realisasi = spse.ScrapePencatatanRealisasi(client, c2, ids)
		case "swakelola":
			results = spse.ScrapeSwakelolaDetails(client, c, ids)
			realisasi = spse.ScrapeSwakelolaRealisasi(client, c2, ids)
		}

		if len(results) == 0 {
			util.PrintVerbose(
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

		if err := spse.ExportToCSV(
			results,
			category,
			); err != nil {
			util.PrintVerbose(
				"[%s] export failed: %v",
				category,
				err,
				)
			continue
		}
	}

	util.PrintVerbose("")
	util.PrintVerbose("All categories completed.")
}
