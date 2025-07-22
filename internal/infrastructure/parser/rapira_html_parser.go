package parser

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"




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
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://rapira.net/exchange/USDT_RUB"),
		chromedp.WaitVisible(`table.table-orders-sell`, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		logger.Log.Errorf("chromedp error: %v", err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		logger.Log.Errorf("failed to parse HTML: %v", err)
		return nil, err
	}

	rows := doc.Find("table.table-orders-sell tbody tr")
	if rows.Length() < 5 {
		logger.Log.Error("not enough bid rows")
		return nil, fmt.Errorf("bid table too short")
	}

	orders := []*domain.Order{}
	rows.EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= 5 {
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

	return orders, nil
}


func FetchRapiraAsk() ([]*domain.Order, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://rapira.net/exchange/USDT_RUB"),
		chromedp.WaitVisible(`table.table-orders-buy`, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		logger.Log.Errorf("chromedp error: %v", err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		logger.Log.Errorf("failed to parse HTML: %v", err)
		return nil, err
	}

	rows := doc.Find("table.table-orders-buy tbody tr")
	if rows.Length() < 5 {
		logger.Log.Error("not enough ask rows")
		return nil, fmt.Errorf("ask table too short")
	}

	orders := []*domain.Order{}
	rows.EachWithBreak(func(i int, s *goquery.Selection) bool {
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
			Side:   domain.SideBuy,
			Source: domain.RapiraSource,
			Pair:   domain.Usdtrub,
		})
		return true
	})
	cleanOrders := make([]*domain.Order, 0)
	for i := len(orders) - 1; i >= len(orders) - 5; i-- {
		cleanOrders = append(cleanOrders, orders[i])
	}
	
	logger.Log.Info("Rapira parsed")
	return cleanOrders, nil
}
