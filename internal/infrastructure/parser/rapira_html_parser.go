package parser

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

//table.table.table-row-dashed.table-orders-buy.gy-1.gs-1.mb-0		красный стакан
//table.table.table-row-dashed.table-orders-sell.gy-1.gs-1			зеленый стакан


func FetchRapiraAsk() (*domain.Order, error) {
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

func FetchRapiraBid() (*domain.Order, error) {
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