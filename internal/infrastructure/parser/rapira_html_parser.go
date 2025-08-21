package parser

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/chromedp/chromedp"
)

//table.table.table-row-dashed.table-orders-buy.gy-1.gs-1.mb-0		красный стакан
//table.table.table-row-dashed.table-orders-sell.gy-1.gs-1			зеленый стакан


func RapiraAsk() (*domain.Order, error) {
	resp, err := http.Get("https://rapira.net/exchange/USDT_RUB")
	if err != nil {
		logger.Log.Errorf("failed to get response: %v", err)
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

	var price float64
	var amount float64

	tableBuy := doc.Find("table.table.table-row-dashed.table-orders-buy.gy-1.gs-1.mb-0")
	lastRow := tableBuy.Last()

	priceStr := strings.ReplaceAll(lastRow.Find("td").Eq(0).Text(),  "\u00a0", "")
	priceStr = strings.ReplaceAll(priceStr,  " ", "")

	amountStr := strings.ReplaceAll(lastRow.Find("td").Eq(1).Text(),  "\u00a0", "")
	amountStr = strings.ReplaceAll(amountStr,  " ", "")

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		logger.Log.Errorf("failed to parse price: %v", err)
		return nil, err
	}
	amount, err = strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logger.Log.Errorf("failed to parse order: %v", err)
		return nil, err
	}
	
	order := domain.Order{
		Price: price,
		Amount: amount,
		Side: domain.SideBuy,
		Source: domain.RapiraSource,
	}
	return &order, nil
}

func RapiraBid() (*domain.Order, error) {
	resp, err := http.Get("https://rapira.net/exchange/USDT_RUB")
	if err != nil {
		logger.Log.Errorf("failed to get response: %v", err)
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

	var price float64	
	var amount float64

	tableBuy := doc.Find("table.table.table-row-dashed.table-orders-sell.gy-1.gs-1")
	firstRow := tableBuy.First()

	priceStr := strings.ReplaceAll(firstRow.Find("td").Eq(0).Text(),  "\u00a0", "")
	priceStr = strings.ReplaceAll(priceStr,  " ", "")

	amountStr := strings.ReplaceAll(firstRow.Find("td").Eq(1).Text(),  "\u00a0", "")
	amountStr = strings.ReplaceAll(amountStr,  " ", "")

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		logger.Log.Errorf("failed to parse price: %v", err)
		return nil, err
	}
	amount, err = strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logger.Log.Errorf("failed to parse order: %v", err)
		return nil, err
	}
	
	order := domain.Order{
		Price: price,
		Amount: amount,
		Side: domain.SideSell,
		Source: domain.RapiraSource,
	}
	return &order, nil
}

func FetchRapiraBid() ([]*domain.Order, error) {
	const (
		maxAttempts  = 3
		opTimeout    = 20 * time.Second
		initialSleep = 500 * time.Millisecond
	)

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var htmlContent string
		err := runOnceWithNewTab(opTimeout,
			chromedp.Navigate("https://rapira.net/exchange/USDT_RUB"),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.Sleep(initialSleep),
			chromedp.ActionFunc(EnsureNoOverlay),
			chromedp.ActionFunc(AcceptCookies),
			chromedp.WaitVisible(`table.table.table-row-dashed.table-orders-sell.gy-1.gs-1 tbody tr`, chromedp.ByQuery),
			chromedp.Sleep(150*time.Millisecond),
			chromedp.OuterHTML(`table.table.table-row-dashed.table-orders-sell.gy-1.gs-1`, &htmlContent, chromedp.ByQuery),
		)

		if err != nil {
			lastErr = err
			logger.Log.Warnf("Rapira Bid attempt %d/%d failed: %v", attempt, maxAttempts, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			lastErr = err
			continue
		}

		rows := doc.Find("table.table-orders-sell tbody tr")
		n := rows.Length()
		if n == 0 {
			lastErr = fmt.Errorf("rapira: bid table empty")
	        continue
		}
		limit := 5
		if n < limit {
			limit = n
		}

		orders := make([]*domain.Order, 0, limit)
		rows.EachWithBreak(func(i int, s *goquery.Selection) bool {
			if i >= limit {
				return false
			}
			tds := s.Find("td")
			if tds.Length() < 3 {
				return true
			}
			price, err1 := sanitizeNumericString(tds.Eq(0).Text())
			amount, err2 := sanitizeNumericString(tds.Eq(1).Text())
			sum, err3 := sanitizeNumericString(tds.Eq(2).Text())
			if err1 != nil || err2 != nil || err3 != nil {
				return true
			}
			orders = append(orders, &domain.Order{
				Price:  price,
				Amount: amount,
				Sum:    sum,
				Side:   domain.SideSell,
				Source: domain.RapiraSource,
				Pair:   domain.Usdtrub,
			})
			return true
		})

		logger.Log.Info("Rapira Bid parsed")
		return orders, nil
	}

	return nil, lastErr
}

func FetchRapiraAsk() ([]*domain.Order, error) {
	const (
		maxAttempts  = 3
		opTimeout    = 20 * time.Second
		initialSleep = 500 * time.Millisecond
	)

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var htmlContent string
		err := runOnceWithNewTab(opTimeout,
			chromedp.Navigate("https://rapira.net/exchange/USDT_RUB"),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.Sleep(initialSleep),
			chromedp.ActionFunc(EnsureNoOverlay),
			chromedp.ActionFunc(AcceptCookies),
			chromedp.WaitVisible(`table.table.table-row-dashed.table-orders-buy.gy-1.gs-1.mb-0 tbody tr`, chromedp.ByQuery),
			chromedp.Sleep(150*time.Millisecond),
			chromedp.OuterHTML(`table.table.table-row-dashed.table-orders-buy.gy-1.gs-1.mb-0`, &htmlContent, chromedp.ByQuery),
		)

		if err != nil {
			lastErr = err
			logger.Log.Warnf("Rapira Ask attempt %d/%d failed: %v", attempt, maxAttempts, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			lastErr = err
			continue
		}

		rows := doc.Find("table.table-orders-buy tbody tr")
		n := rows.Length()
		if n == 0 {
			lastErr = fmt.Errorf("rapira: bid table empty")
	        continue
		}
		all := make([]*domain.Order, 0, n)
		rows.Each(func(i int, s *goquery.Selection) {
			tds := s.Find("td")
			if tds.Length() < 3 {
				return
			}
			price, err1 := sanitizeNumericString(tds.Eq(0).Text())
			amount, err2 := sanitizeNumericString(tds.Eq(1).Text())
			sum, err3 := sanitizeNumericString(tds.Eq(2).Text())
			if err1 != nil || err2 != nil || err3 != nil {
				return
			}
			all = append(all, &domain.Order{
				Price:  price,
				Amount: amount,
				Sum:    sum,
				Side:   domain.SideBuy,
				Source: domain.RapiraSource,
				Pair:   domain.Usdtrub,
			})
		})

		if len(all) == 0 {
			lastErr = fmt.Errorf("ask table parsed empty")
			continue
		}

		limit := 5
		if len(all) < limit {
			limit = len(all)
		}
		clean := make([]*domain.Order, 0, limit)
		for i := len(all)-1; i >= len(all)-limit; i-- {
			clean = append(clean, all[i])
		}

		logger.Log.Info("Rapira Ask parsed")
		return clean, nil
	}

	return nil, lastErr
}
