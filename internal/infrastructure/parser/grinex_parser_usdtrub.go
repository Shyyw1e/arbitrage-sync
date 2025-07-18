package parser

import (
	"errors"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

func GrinexAskUSDTRub() (*domain.Order, error) {
	table, err := PreFetchGrinex("usdtrub" ,"ask_orders_panel")
	if err != nil {
		logger.Log.Errorf("failed to prefetch Grinex html: %v", err)
		return nil, err
	}
	row := table.First()

	priceStr,	ok := row.Attr("data-price")
	if !ok {
		logger.Log.Error("empty price")
		return nil, errors.New("empty price")
	}
	amountStr,	ok := row.Attr("data-volume")
	if !ok {
		logger.Log.Error("empty amount")
		return nil, errors.New("empty amount")
	}
	
	price, err := sanitizeNumericString(priceStr)
	if err != nil {
		logger.Log.Errorf("failed to parse price: %v", err)
		return nil, err
	}
	amount, err := sanitizeNumericString(amountStr)
	if err != nil {
		logger.Log.Errorf("failed to parse order: %v", err)
		return nil, err
	}

	order := domain.Order{
		Price: 	price,
		Amount: amount,
		Side: 	domain.SideBuy,
		Source: domain.GrinexSource,
		Pair: 	domain.Usdtrub,
	}

	return &order, nil
}

func GrinexBidUSDTRub() (*domain.Order, error) {
	table, err := PreFetchGrinex("usdtrub" ,"bid_orders_panel")
	if err != nil {
		logger.Log.Errorf("failed to prefetch Grinex html: %v", err)
		return nil, err
	}
	row := table.First()

	priceStr,	ok := row.Attr("data-price")
	if !ok {
		logger.Log.Error("empty price")
		return nil, errors.New("empty price")
	}
	amountStr,	ok := row.Attr("data-volume")
	if !ok {
		logger.Log.Error("empty amount")
		return nil, errors.New("empty amount")
	}

	priceStr = strings.ReplaceAll(priceStr, "\u00a0", "")
	priceStr = strings.ReplaceAll(priceStr, " ", "")
	amountStr = strings.ReplaceAll(amountStr, "\u00a0", "")
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		logger.Log.Errorf("failed to parse price: %v", err)
		return nil, err
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logger.Log.Errorf("failed to parse order: %v", err)
		return nil, err
	}

	order := domain.Order{
		Price: 	price,
		Amount: amount,
		Side: 	domain.SideSell,
		Source: domain.GrinexSource,
		Pair: 	domain.Usdtrub,
	}

	return &order, nil
}

func FetchGrinexAskUSDTRub() ([]*domain.Order, error) {
	table, err := PreFetchGrinex("usdtrub" ,"ask_orders_panel")
	if err != nil {
		logger.Log.Errorf("failed to prefetch Grinex html: %v", err)
		return nil, err
	}
	
	orders := make([]*domain.Order, 0)
	count := 0 
	table.Each(func(i int, s *goquery.Selection) {
		if count < 5 {
			priceStr,	ok := s.Attr("data-price")
			if !ok {
				logger.Log.Error("empty price")
				return 
			}
			amountStr,	ok := s.Attr("data-volume")
			if !ok {
				logger.Log.Error("empty amount")
				return 
			}
			
			price, err := sanitizeNumericString(priceStr)
			if err != nil {
				logger.Log.Errorf("failed to parse price: %v", err)
				return 
			}
			amount, err := sanitizeNumericString(amountStr)
			if err != nil {
				logger.Log.Errorf("failed to parse order: %v", err)
				return 
			}

			order := domain.Order{
				Price: price,
				Amount: amount,
				Side: domain.SideBuy,
				Source: domain.GrinexSource,
				Pair: domain.Usdtrub,
			}

			orders = append(orders, &order)
			count++
		} else {
			return 
		}
		
	})

	return orders, nil
}

func FetchGrinexBidUSDTRub() ([]*domain.Order, error) {
	table, err := PreFetchGrinex("usdtrub" ,"bid_orders_panel")
	if err != nil {
		logger.Log.Errorf("failed to prefetch Grinex html: %v", err)
		return nil, err
	}
	
	orders := make([]*domain.Order, 0)
	count := 0 
	table.Each(func(i int, s *goquery.Selection) {
		if count < 5 {
			priceStr,	ok := s.Attr("data-price")
			if !ok {
				logger.Log.Error("empty price")
				return 
			}
			amountStr,	ok := s.Attr("data-volume")
			if !ok {
				logger.Log.Error("empty amount")
				return 
			}
			
			price, err := sanitizeNumericString(priceStr)
			if err != nil {
				logger.Log.Errorf("failed to parse price: %v", err)
				return 
			}
			amount, err := sanitizeNumericString(amountStr)
			if err != nil {
				logger.Log.Errorf("failed to parse order: %v", err)
				return 
			}

			order := domain.Order{
				Price: price,
				Amount: amount,
				Side: domain.SideSell,
				Source: domain.GrinexSource,
				Pair: domain.Usdtrub,
			}

			orders = append(orders, &order)
			count++
		} else {
			return 
		}
		
	})

	return orders, nil
}