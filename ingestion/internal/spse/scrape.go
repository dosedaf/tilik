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
	"encoding/json"

	"github.com/gocolly/colly/v2"
	"ingestion/internal/spse/category"
	"ingestion/internal/spse/model"
	"ingestion/util"
)

type ScrapeContext struct {
	Pemda string
	Year string
}

func (s *SPSEScraper) Scrape(pemdas []string, years []string) error {
	for _, pemda := range pemdas {
		for _, year := range years {
			ctx := ScrapeContext{
				Pemda: pemda,
				Year: year,
			}

			for _, cfg := range category.AllConfigs() {
				if err := s.scrapeCategory(ctx, cfg); err != nil {
					util.PrintVerbose("[%s/%s] failed: %v",ctx.Pemda, cfg.Category, err)
					return err
				}
			}
		}
	}

	return nil
}

func (s *SPSEScraper) scrapeCategory(ctx ScrapeContext, cfg category.ScraperConfig) error {
	util.PrintVerbose("CATEGORY: %s", cfg.Category)

	token, err := s.GetToken(
		ctx.Pemda,
		cfg,
		)

	if err != nil {
		util.PrintVerbose("[%s] FATAL: %v", cfg.Category, err)
		return err
	}

	ids, err := s.FetchIDs(
		token,
		ctx.Pemda,
		cfg,
		)

	if err != nil {
		util.PrintVerbose("[%s] FATAL: %v", cfg.Category, err)
		return err
	}

	if limit := model.ScrapeLimits[cfg.Category]; limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	} else if limit == 0{
		ids = ids[:0] 
	}

	if len(ids) == 0 {
		util.PrintVerbose("[%s] no records found", cfg.Category)
		return nil
	}

	c := s.NewCollector(cfg.Category, ctx.Pemda)
	c2 := s.NewCollector(cfg.Category, ctx.Pemda)

	var results []model.Paket
	var pemenangBerkontrak map[string]string
	var realisasi map[string][]model.Realisasi

	switch cfg.Category {
	case "tender":
		results = s.ScrapeDetails(ctx, c, ids, cfg)
		pemenangBerkontrak = s.ScrapePemenangBerkontrak(ctx, c2, ids, cfg)
	case "nontender":
		results = s.ScrapeDetails(ctx, c, ids, cfg)
		pemenangBerkontrak = s.ScrapePemenangBerkontrak(ctx, c2, ids, cfg)
	case "pencatatan":
		results = s.ScrapeDetails(ctx, c, ids, cfg)
		realisasi = s.ScrapeRealisasi(ctx, c2, ids, cfg)
	case "swakelola":
		results = s.ScrapeDetails(ctx, c, ids, cfg)
		realisasi = s.ScrapeRealisasi(ctx, c2, ids, cfg)
	}

	if len(results) == 0 {
		util.PrintVerbose("[%s] no package details scraped", cfg.Category)
	}

	switch cfg.Category {
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

	if err := s.ExportToCSV(
		results,
		cfg.Category,
		); err != nil {
		util.PrintVerbose("[%s] export failed: %v", cfg.Category, err)
	}

	return nil
}

func (s *SPSEScraper) FetchIDs(
	token string,
	pemda string,
	cfg category.ScraperConfig,
) ([]string, error) {
	apiURL := GetPath(cfg.Category, pemda, "", "dt")

	if apiURL == "" {
		return nil, fmt.Errorf("invalid DT path for category %s", cfg.Category)
	}

	apiURL += "?tahun=" + url.QueryEscape(model.Year)

	util.PrintVerbose("[%s] POST %s", cfg.Category, apiURL)

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

	resp, err := s.Client.Do(req)
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
		cfg.Category,
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
		cfg.Category,
		len(ids),
		)

	return ids, nil
}

func (s *SPSEScraper) ScrapeDetails(ctx ScrapeContext, c *colly.Collector, ids []string, cfg category.ScraperConfig) []model.Paket {
	var results []model.Paket
	var mu sync.Mutex
	var scraped atomic.Int64
	wsRegex := regexp.MustCompile(`\s+`)

	c.OnHTML("html", func(e *colly.HTMLElement) {
		detail := cfg.InitDetail(e.Request.URL.String())

		if detail.Kode == "" {
			detail.Kode = cfg.ExtractDetailKode(e.Request.URL)
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
		targetURL := GetPath(cfg.Category, ctx.Pemda, id, "pengumuman")
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

func (s *SPSEScraper)  ScrapePemenangBerkontrak(ctx ScrapeContext, c *colly.Collector, ids []string, cfg category.ScraperConfig) map[string]string {
	pemenang := make(map[string]string)
	var mu sync.Mutex

	c.OnHTML("html", func(e *colly.HTMLElement) {
		kode := cfg.ExtractEvaluasiKode(e.Request.URL)

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
			ctx.Pemda,
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


func (s *SPSEScraper) ScrapeRealisasi(ctx ScrapeContext, c *colly.Collector, ids []string, cfg category.ScraperConfig) map[string][]model.Realisasi {
	realisasiData := make(map[string][]model.Realisasi)
	var mu sync.Mutex

	c.OnHTML("html", func(e *colly.HTMLElement) {
		selector := ".bs-callout-info + table.table-sm tr:has(td[align=\"center\"])"

		kode := cfg.ExtractEvaluasiKode(e.Request.URL)

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
			ctx.Pemda,
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
