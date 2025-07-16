package parser

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
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

	

	tableBuy := doc.Find("table.table.table-row-dashed.table-orders-sell.gy-1.gs-1")
	count := 0
	orders := make([]*domain.Order, 0)
	tableBuy.Each(func(i int, s *goquery.Selection) {
		if count < 5 {
			tds := s.Find("td")
			if tds.Length() < 3 {
				return
			}

			priceStr := tds.Eq(0).Text()
			amountStr := tds.Eq(1).Text()
			sumStr := tds.Eq(2).Text()
			price, err := sanitizeNumericString(priceStr)
			if err != nil {
				logger.Log.Errorf("failed to convert string price: %v", err)
				return 
			}
			amount, err := sanitizeNumericString(amountStr)
			if err != nil {
				logger.Log.Errorf("failed to convert string amount: %v", err)
				return 
			}
			sum, err := sanitizeNumericString(sumStr)
			if err != nil {
				logger.Log.Errorf("failed to convert string sum: %v", err)
				return 
			}

			orders = append(orders, &domain.Order{
				Price: price,
				Amount: amount,
				Sum: sum,
				Side: domain.SideSell,
				Source: domain.RapiraSource,
				Pair: domain.Usdtrub,
			})
			count++
		} else {
			return 
		}

		
	})

	return orders, nil
}

func FetchRapiraAsk() ([]*domain.Order, error) {
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


	tableBuy := doc.Find("table.table.table-row-dashed.table-orders-buy.gy-1.gs-1.mb-0").Nodes
	if len(tableBuy) < 5 {
		err := errors.New("invalid table (length<5)")
		logger.Log.Errorf("failed to find table: %v", err)
		return nil, err
	}
	rows := tableBuy[len(tableBuy) - 5:]

	orders := make([]*domain.Order, 0)

	for _, row := range rows {
		sel := goquery.NewDocumentFromNode(row)
		tds := sel.Find("td")
		if tds.Length() < 3 {
			continue
		}
		priceStr := tds.Eq(0).Text()
		amountStr := tds.Eq(1).Text()
		sumStr := tds.Eq(2).Text()
		price, err := sanitizeNumericString(priceStr)
		if err != nil {
			logger.Log.Errorf("failed to convert string price: %v", err)
			continue 
		}
		amount, err := sanitizeNumericString(amountStr)
		if err != nil {
			logger.Log.Errorf("failed to convert string amount: %v", err)
			continue
		}
		sum, err := sanitizeNumericString(sumStr)
		if err != nil {
			logger.Log.Errorf("failed to convert string sum: %v", err)
			continue
		}

		orders = append(orders, &domain.Order{
			Price: price,
			Amount: amount,
			Sum: sum,
			Side: domain.SideBuy,
			Source: domain.RapiraSource,
			Pair: domain.Usdtrub,
		})

	}

	return orders, nil
}