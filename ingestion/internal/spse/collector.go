package spse 

import (
	"fmt"
	"net/url"
	"time"

	"ingestion/internal/spse/model"
	"github.com/gocolly/colly/v2"
)

func (s *SPSEScraper) NewCollector(category string, pemda string) (*colly.Collector){
	c := colly.NewCollector(
		colly.AllowedDomains("spse.inaproc.id"),
		colly.UserAgent(model.UserAgent),
		colly.Async(true),
		)

	baseURLParsed, err := url.Parse(model.BaseURL)

	if err == nil && s.Client.Jar != nil {
		cookies := s.Client.Jar.Cookies(baseURLParsed)

		if len(cookies) > 0 {
			_ = c.SetCookies(model.BaseURL, cookies)
		}
	}

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*spse.inaproc.id*",
		Parallelism: 4,
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
			GetPath(category, pemda, "", "portal"),
			)
	})

	return c
}
