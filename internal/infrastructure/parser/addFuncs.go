package parser

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/chromedp/chromedp"
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

func PreFetchGrinex(endpoint string, panelClass string) (*goquery.Selection, error) {
	const maxAttempts = 3
	const delayBetweenAttempts = 3 * time.Second

	var htmlContent string
	url := "https://grinex.io/trading/" + endpoint

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := chromedp.NewContext(context.Background())
		timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 12 * time.Second)

		logger.Log.Infof("Starting chromedp session for: %s [%s]", endpoint, panelClass)
		err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(url),
			chromedp.Sleep(6*time.Second),
			chromedp.WaitReady("div.order_book_holder", chromedp.ByQuery),
			chromedp.WaitVisible("div."+panelClass+" table", chromedp.ByQuery),
			chromedp.OuterHTML("html", &htmlContent),
		)
		cancelTimeout()
		cancel()

		if err == nil {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
			if err != nil {
				logger.Log.Errorf("goquery parse error: %v", err)
				return nil, err
			}

			mainDiv := doc.Find("div.order_book_holder")
			sideDiv := mainDiv.Find("div." + panelClass)
			tableSelection := sideDiv.Find("table")
			if tableSelection.Length() == 0 {
				logger.Log.Error("invalid table (not found)")
				return nil, errors.New("invalid table")
			}

			tableRows := tableSelection.Find("tbody tr")
			if tableRows.Length() == 0 {
				logger.Log.Error("empty orderbook rows")
				return nil, errors.New("empty orderbook")
			}

			return tableRows, nil
		}

		logger.Log.Warnf("Attempt %d: error loading %s [%s] — %v", attempt, endpoint, panelClass, err)
		time.Sleep(delayBetweenAttempts)
	}

	return nil, fmt.Errorf("failed to load Grinex %s (%s) after %d attempts", endpoint, panelClass, maxAttempts)
}

/*
	func GetTaskHandler()

*/