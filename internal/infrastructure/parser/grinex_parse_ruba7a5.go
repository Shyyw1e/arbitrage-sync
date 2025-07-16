package parser

import (
	"errors"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

func GrinexAskRubA7A5() (*domain.Order, error) {
	table, err := PreFetchGrinex("a7a5rub" ,"ask_orders_panel")
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
		Price:	price,
		Amount: amount,
		Side: 	domain.SideBuy,
		Source: domain.GrinexSource,
		Pair: 	domain.A7a5rub,
	}

	return &order, nil
}

func GrinexBidRubA7A5() (*domain.Order, error) {
	table, err := PreFetchGrinex("a7a5rub" ,"bid_orders_panel")
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
		Price: price,
		Amount: amount,
		Side: domain.SideSell,
		Source: domain.GrinexSource,
		Pair: domain.A7a5rub,
	}

	return &order, nil
}

func FetchGrinexAskA7A5Rub() ([]*domain.Order, error) {
	table, err := PreFetchGrinex("a7a5rub" ,"ask_orders_panel")
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
				Pair: domain.A7a5rub,
			}

			orders = append(orders, &order)
			count++
		} else {
			return 
		}
		
	})

	return orders, nil
}

func FetchGrinexBidA7A5Rub() ([]*domain.Order, error) {
	table, err := PreFetchGrinex("a7a5rub" ,"bid_orders_panel")
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
				Pair: domain.A7a5rub,
			}

			orders = append(orders, &order)
			count++
		} else {
			return 
		}
		
	})

	return orders, nil
}