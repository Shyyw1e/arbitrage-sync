package parser

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

func sanitizeNumericString(s string) (float64, error) {
    s = strings.ReplaceAll(s, "\u00a0", "")
    s = strings.ReplaceAll(s, " ", "")
	res, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Log.Errorf("failed to parse float64: %v", err)
		return -1, err
	}

	return res, nil
}

func PreFetchGrinexUSDTRub(panelClass string) (*goquery.Selection, error) {
	resp, err := http.Get("https://grinex.io/trading/usdtrub")
	if err != nil {
		logger.Log.Errorf("failed to make response: %v", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("invalid status cod (not 200)")
		return nil, err
	}
	if resp.Body == nil {
		logger.Log.Error("empty response body")
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Log.Errorf("failed to make document from resp.Body: %v", err)
		return nil, err
	}

	mainDiv := doc.Find("div.order_book_holder")
	sideDiv := mainDiv.Find("div." + panelClass)
	tableSelection := sideDiv.Find("table")
	if tableSelection.Length() <= 0 {
		logger.Log.Error("invalid table")
		return nil, errors.New("invalid table")
	}
	tableSelection = tableSelection.Find("tbody tr")
	return tableSelection, nil
}