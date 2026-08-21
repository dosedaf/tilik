package spse

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
	"net/http/cookiejar"
	"os"

	"ingestion/internal/spse/model"
	"ingestion/internal/spse/category"
	"ingestion/util"
)

type SPSEScraper struct {
	Client *http.Client
}

func NewSPSEScraper() *SPSEScraper {
	jar, err := cookiejar.New(nil)

	if err != nil {
		util.PrintVerbose("FATAL: %v", err)
		os.Exit(1)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}


	return &SPSEScraper{
		Client: client,
	}
}

func (s *SPSEScraper) GetToken(pemda string, cfg category.ScraperConfig) (string, error) {
	reqURL := GetPath(cfg.Category, pemda, "", "portal")

	if reqURL == "" {
		return "", fmt.Errorf("invalid category: %s", cfg.Category)
	}

	util.PrintVerbose("[%s] GET %s", cfg.Category, reqURL)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", model.UserAgent)

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	util.PrintVerbose(
		"[%s] HTTP/1.1 %d %s",
		cfg.Category,
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
		return "", fmt.Errorf("authenticityToken not found for category %s", cfg.Category)
	}

	util.PrintVerbose("[%s] authenticityToken found", cfg.Category)

	return matches[1], nil
}

