package spse

import (
	"net/http"
	"fmt"
	"time"
	"regexp"
	"strings"
	"sync/atomic"
	"sync"
	"net/url"

	"github.com/gocolly/colly/v2"
	"ingestion/internal/spse/category"
	"ingestion/internal/spse/model"
	"ingestion/util"
)

func extractKode(path, prefix string) string {
    idx := strings.Index(path, prefix)
    if idx == -1 {
        return ""
    }

    remaining := path[idx+len(prefix):]

    return strings.Split(remaining, "/")[0]
}

func NewScraper(client *http.Client, category string) (*colly.Collector){
	c := colly.NewCollector(
		colly.AllowedDomains("spse.inaproc.id"),
		colly.UserAgent(model.UserAgent),
		colly.Async(true),
		)

	baseURLParsed, err := url.Parse(model.BaseURL)

	if err == nil && client.Jar != nil {
		cookies := client.Jar.Cookies(baseURLParsed)

		if len(cookies) > 0 {
			_ = c.SetCookies(model.BaseURL, cookies)
		}
	}

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*spse.inaproc.id*",
		Parallelism: 8,
		Delay:       time.Second / 2,
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
			GetPath(category, model.Pemda, "", "portal"),
			)
	})

	return c
}

func ScrapeDetails(client *http.Client, c *colly.Collector, ids []string, cfg category.ScraperConfig) []model.Paket {
	var results []model.Paket
	var mu sync.Mutex
	var scraped atomic.Int64
	wsRegex := regexp.MustCompile(`\s+`)

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := cfg.InitDetail(e.Request.URL.String())
		if detail.Kode == "" {
			detail.Kode = extractKode(e.Request.URL.Path, cfg.KodePrefix.Detail)
		}

		e.ForEach("table.table tr", func(_ int, row *colly.HTMLElement) {
			rawKey := strings.TrimSpace(row.ChildText("th"))
			rawVal := strings.TrimSpace(row.ChildText("td"))
			keyLower := strings.ToLower(wsRegex.ReplaceAllString(rawKey, " "))
			val := wsRegex.ReplaceAllString(rawVal, " ")
			if keyLower == "" || val == "" {
				return
			}
			for _, rule := range cfg.FieldRules {
				if rule.Match(keyLower) {
					rule.Handle(&detail, val)
					break
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
		util.PrintVerbose("[%s] <= 200 OK %s %d/%d", cfg.Category, detail.Kode, current, len(ids))
	})

	c.OnError(func(r *colly.Response, err error) {
		util.PrintVerbose("[%s] <= %d ERROR %v", cfg.Category, r.StatusCode, err)
	})

	for _, id := range ids {
		targetURL := GetPath(cfg.Category, model.Pemda, id, "pengumuman")
		if targetURL == "" {
			util.PrintVerbose("[%s] skipping ID %s: no %s path", cfg.Category, id, "pengumuman")
			continue
		}
		if err := c.Visit(targetURL); err != nil {
			util.PrintVerbose("[%s] failed to visit %s: %v", cfg.Category, targetURL, err)
		}
	}
	c.Wait()
	return results
}

func ScrapeTenderDetails(client *http.Client, c *colly.Collector, ids []string) []model.Paket {
	return ScrapeDetails(client, c, ids, category.TenderConfig())
}

func ScrapeNonTenderDetails(client *http.Client, c *colly.Collector, ids []string) []model.Paket {
	return ScrapeDetails(client, c, ids, category.NonTenderConfig()) // same shape, different rules/prefix
}

func ScrapePencatatanDetails(client *http.Client, c *colly.Collector, ids []string) []model.Paket {
	return ScrapeDetails(client, c, ids, category.PencatatanConfig()) // same shape, different rules/prefix
}

func ScrapeSwakelolaDetails(client *http.Client, c *colly.Collector, ids []string) []model.Paket {
	return ScrapeDetails(client, c, ids, category.SwakelolaConfig()) // same shape, different rules/prefix
}

func ScrapePemenangBerkontrak(client *http.Client, c *colly.Collector, ids []string, cfg category.ScraperConfig) map[string]string {
	pemenang := make(map[string]string)
	var mu sync.Mutex

	c.OnHTML("html", func(e *colly.HTMLElement) {
		urlPath := e.Request.URL.Path
		kode := extractKode(urlPath, cfg.KodePrefix.Evaluasi)

		if cfg.Category == "pencatatan" {
			kode = e.Request.URL.Query().Get("id")
		}

		detailTag := e.DOM.Find("table.table tr:last-child tr td:first-child")

		if detailTag.Length() > 0 {
			mu.Lock()
			pemenang[kode] = strings.TrimSpace(detailTag.Text())
			mu.Unlock()
			return 
		}

		mu.Lock()
		pemenang[kode] = "Tidak ada"
		mu.Unlock()
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("[%s] error: %v\n", cfg.Category, err)
	})

	for _, id := range ids {
		targetURL := GetPath(
			cfg.Category,
			model.Pemda,
			id,
			"pemenang_berkontrak",
		)

		if targetURL == "" {
			util.PrintVerbose(
				"[%s] skipping ID %s: no pemenang berkontrak path",
				cfg.Category,
				id,
			)
			continue
		}

		if err := c.Visit(targetURL); err != nil {
			util.PrintVerbose(
				"[%s] failed to visit %s: %v",
				cfg.Category,
				targetURL,
				err,
			)
		}
	}

	c.Wait()

	return pemenang

}

func ScrapeTenderPemenangBerkontrak(client *http.Client, c *colly.Collector, ids []string) map[string]string {
	return ScrapePemenangBerkontrak(client, c, ids, category.TenderConfig())
}

func ScrapeNonTenderPemenangBerkontrak(client *http.Client, c *colly.Collector, ids []string) map[string]string {
	return ScrapePemenangBerkontrak(client, c, ids, category.NonTenderConfig())
}

func ScrapeRealisasi(client *http.Client, c *colly.Collector, ids []string, cfg category.ScraperConfig) map[string][]model.Realisasi {
	realisasiData := make(map[string][]model.Realisasi)
	var mu sync.Mutex

	c.OnHTML("html", func(e *colly.HTMLElement) {
		selector := ".bs-callout-info + table.table-sm tr:has(td[align=\"center\"])"

		urlPath := e.Request.URL.Path
		kode := extractKode(urlPath, cfg.KodePrefix.Evaluasi)

		if cfg.Category == "pencatatan" {
			kode = e.Request.URL.Query().Get("id")
		}

		var realisasi []model.Realisasi

		if e.DOM.Find(selector).Length() == 0  {
			mu.Lock()
			realisasiData[kode] = []model.Realisasi{}
			mu.Unlock()
			return
		}

		e.ForEach(selector, func(_ int, e *colly.HTMLElement) {
			jenis := e.ChildText("td:nth-child(2)")
			nilai := e.ChildText("td:nth-child(3)")
			tanggal := e.ChildText("td:nth-child(4)")
			
			// DUMB QUICK FIX
			// assign emptry string to tanggal if err is found
			// FIX THIS LATER
			var parsedTanggal time.Time

			if tanggal != "" {
				var err error
				parsedTanggal, err = time.Parse("02-01-2006", tanggal)

				if err != nil {
					fmt.Printf("[%s] error parsing date:%s \n", kode, err)
				}
			}

			dateOnly := time.Date(
				parsedTanggal.Year(),
				parsedTanggal.Month(),
				parsedTanggal.Day(),
				0,0,0,0,
				parsedTanggal.Location(),
				)

			r := model.Realisasi{
				Jenis: jenis,
				Nilai: nilai,
				Tanggal: dateOnly,
			}

			realisasi = append(realisasi, r)
		})

		mu.Lock()
		realisasiData[kode] = realisasi
		mu.Unlock()
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("[%s] error: %v\n", cfg.Category, err)
	})

	for _, id := range ids {
		targetURL := GetPath(
			cfg.Category,
			model.Pemda,
			id,
			"pemenang_berkontrak",
		)

		if targetURL == "" {
			util.PrintVerbose(
				"[%s] skipping ID %s: no pemenang path",
				cfg.Category,
				id,
			)
			continue
		}

		if err := c.Visit(targetURL); err != nil {
			util.PrintVerbose(
				"[%s] failed to visit %s: %v",
				cfg.Category,
				targetURL,
				err,
			)
		}
	}

	c.Wait()

	return realisasiData
}

func ScrapePencatatanRealisasi(client *http.Client, c *colly.Collector, ids []string) map[string][]model.Realisasi{
	return ScrapeRealisasi(client, c, ids, category.PencatatanConfig())
}

func ScrapeSwakelolaRealisasi(client *http.Client, c *colly.Collector, ids []string) map[string][]model.Realisasi{
	return ScrapeRealisasi(client, c, ids, category.SwakelolaConfig())
}
