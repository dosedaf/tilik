package spse

import (
	"fmt"

	"ingestion/internal/spse/model"
)

func GetPath(category, pemda, kode, page string) string {
	paths, ok := model.CategoryPaths[category]
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

	return fmt.Sprintf("%s/%s%s", model.BaseURL, pemda, path)
}

