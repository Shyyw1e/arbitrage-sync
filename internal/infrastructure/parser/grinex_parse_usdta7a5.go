package parser

import (
	"errors"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

func GrinexAskUSDTA7A5() (*domain.Order, error) {
	table, err := PreFetchGrinex("usdta7a5" ,"ask_orders_panel")
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
		Pair: 	domain.Usdta7a5,
	}

	return &order, nil
}

func GrinexBidUSDTA7A5() (*domain.Order, error) {
	table, err := PreFetchGrinex("usdta7a5" ,"bid_orders_panel")
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
		Pair: domain.Usdta7a5,
	}

	return &order, nil
}

func FetchGrinexAskUSDTA7A5() ([]*domain.Order, error) {
	table, err := PreFetchGrinex("usdta7a5" ,"ask_orders_panel")
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
				Pair: domain.Usdta7a5,
			}

			orders = append(orders, &order)
			count++
		} else {
			return 
		}
		
	})

	

	

	return orders, nil
}

func FetchGrinexBidUSDTA7A5() ([]*domain.Order, error) {
	table, err := PreFetchGrinex("usdta7a5" ,"bid_orders_panel")
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
				Pair: domain.Usdta7a5,
			}

			orders = append(orders, &order)
			count++
		} else {
			return 
		}
		
	})

	

	

	return orders, nil
}




/*
	func DetectAS() {
		rapiraRed, err := FetchRapiraAsk()
		if err != nil {
			logger.Log.Errorf("failed to fetch Rapira ask table: %v", err)
			return nil, err
		}
		rapiraGreen, err := FetchRapiraBid()
		if err != nil {
			logger.Log.Errorf("failed to fetch Rapira bid table: %v", err)
			return nil, err
		}
		GrinexUSDTRUBRed, err := FetchGrinexAskUSDTRub()
		if err != nil {
			logger.Log.Errorf("failed to fetch Grinex USDT/RUB ask table: %v", err)
			return nil, err
		}
		GrinexUSDTRUBGreen, err := FetchGrinexBidUSDTRub()
		if err != nil {
			logger.Log.Errorf("failed to fetch Grinex USDT/RUB bid table: %v", err)
			return nil, err
		}
		GrinexUSDTA7A5Red, err := FetchGrinexAskUSDTA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch Grinex USDT/A7A5 ask table: %v", err)
			return nil, err
		}
		GrinexUSDTA7A5Green, err := FetchGrinexBidUSDTA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch Grinex USDT/A7A5 bid table: %v", err)
			return nil, err
		}
		GrinexRubA7A5Red, err := FetchGrinexAskRubA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch Grinex A7A5/RUB ask table: %v", err)
			return nil, err
		}
		GrinexRubA7A5Green, err := FetchGrinexBidRubA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch Grinex A7A5/RUB bid table: %v", err)
			return nil, err
		}






	}












*/