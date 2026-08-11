package main

import (
	"net/http"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
)

// new collector each category?
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

func newScraper(client *http.Client, category string) (*colly.Collector){
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

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*spse.inaproc.id*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	if err != nil {
		fmt.Println("error")
	}

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
	})

	return c
}

func scrapeTenderDetails(client *http.Client, c *colly.Collector, ids []string) []Paket {
	category := "tender"
	var results []Paket
	var mu sync.Mutex
	var scraped atomic.Int64

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := Paket{
			Kategori: category,
			URL:      e.Request.URL.String(),
			Tender:	&TenderDetail{},
		}

		urlPath := e.Request.URL.Path

		detail.Kode = extractKode(urlPath, "/lelang/")

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
			case strings.EqualFold(keyLower, "kode tender"):
				detail.Kode = val
				fmt.Println(val)

			case strings.EqualFold(keyLower, "nama tender"):
				detail.Nama = val

			case strings.Contains(keyLower, "k/l/pd"):
				detail.Instansi = val

			case strings.EqualFold(keyLower, "satuan kerja"):
				detail.Satker = val

				if val == "1.02.0.00.0.00.01.0000" {
					detail.Satker = "Dinas Kesehatan"
				}

			case strings.EqualFold(keyLower, "jenis pengadaan"):
				detail.Tender.JenisPengadaan = val

			case strings.EqualFold(keyLower, "metode pengadaan"):
				detail.Tender.MetodePengadaan = val

			case strings.Contains(keyLower, "tahun anggaran"):
				detail.Tahun = val

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
					detail.Pagu = numbers[0]
				}

				if len(numbers) >= 2 {
					detail.Tender.HPS = numbers[1]
				}

			case strings.Contains(keyLower, "lokasi"):
				detail.Tender.Lokasi = val

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

func scrapeNonTenderDetails(client *http.Client, c *colly.Collector, ids []string) []Paket {
	category := "nontender"
	var results []Paket
	var mu sync.Mutex
	var scraped atomic.Int64

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := Paket{
			Kategori: category,
			URL:      e.Request.URL.String(),
			NonTender:	&NonTenderDetail{},
		}

		urlPath := e.Request.URL.Path

		detail.Kode = extractKode(urlPath, "/nontender/")

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
			case strings.EqualFold(keyLower, "kode paket"):
				detail.Kode = val

			case strings.EqualFold(keyLower, "nama paket"):
				detail.Nama = val

			case strings.Contains(keyLower, "k/l/pd"):
				detail.Instansi = val

			case strings.EqualFold(keyLower, "satuan kerja"):
				detail.Satker = val

				if val == "1.02.0.00.0.00.01.0000" {
					detail.Satker = "Dinas Kesehatan"
				}

			case strings.EqualFold(keyLower, "jenis pengadaan"):
				detail.NonTender.JenisPengadaan = val

			case strings.EqualFold(keyLower, "metode pengadaan"):
				detail.NonTender.MetodePengadaan = val

			case strings.Contains(keyLower, "tahun anggaran"):
				detail.Tahun = val

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
					detail.Pagu = numbers[0]
				}

				if len(numbers) >= 2 {
					detail.NonTender.HPS = numbers[1]
				}

			case strings.Contains(keyLower, "lokasi"):
				detail.NonTender.Lokasi = val
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

func scrapePencatatanDetails(client *http.Client, c *colly.Collector, ids []string) []Paket {
	category := "pencatatan"
	var results []Paket
	var mu sync.Mutex
	var scraped atomic.Int64

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := Paket{
			Kategori: category,
			URL:      e.Request.URL.String(),
			Pencatatan:	&PencatatanDetail{},
		}

		detail.Kode = e.Request.URL.Query().Get("id")

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
			case strings.EqualFold(keyLower, "kode paket"):
				detail.Kode = val

			case strings.EqualFold(keyLower, "nama paket"):
				detail.Nama = val

			case strings.Contains(keyLower, "k/l/pd"):
				detail.Instansi = val

			case strings.EqualFold(keyLower, "satuan kerja"):
				detail.Satker = val

				if val == "1.02.0.00.0.00.01.0000" {
					detail.Satker = "Dinas Kesehatan"
				}

			case strings.EqualFold(keyLower, "jenis pengadaan"):
				detail.Pencatatan.JenisPengadaan = val

			case strings.EqualFold(keyLower, "metode pengadaan"):
				detail.Pencatatan.MetodePengadaan = val

			case strings.Contains(keyLower, "tahun anggaran"):
				detail.Tahun = val

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
					detail.Pagu = numbers[0]
				}
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

func scrapeSwakelolaDetails(client *http.Client, c *colly.Collector, ids []string) []Paket {
	category := "swakelola"
	var results []Paket
	var mu sync.Mutex
	var scraped atomic.Int64

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := Paket{
			Kategori: category,
			URL:      e.Request.URL.String(),
			Swakelola:	&SwakelolaDetail{},
		}

		urlPath := e.Request.URL.Path

		detail.Kode = extractKode(urlPath, "/swakelola/")

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
			case strings.EqualFold(keyLower, "kode swakelola"):
				detail.Kode = val

			case strings.EqualFold(keyLower, "nama swakelola"):
				detail.Nama = val

			case strings.Contains(keyLower, "k/l/pd"):
				detail.Instansi = val

			case strings.EqualFold(keyLower, "satuan kerja"):
				detail.Satker = val

				if val == "1.02.0.00.0.00.01.0000" {
					detail.Satker = "Dinas Kesehatan"
				}

			case strings.EqualFold(keyLower, "tipe pelaksana swakelola"):
				detail.Swakelola.TipePelaksana = val

			case strings.Contains(keyLower, "tahun anggaran"):
				detail.Tahun = val

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
					detail.Pagu = numbers[0]
				}
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
