package usecase

import (
	"errors"

	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/parser"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

func Convert(from, to domain.Direction, amount float64) (float64, error) {
	if from + "/" + to == "USDT/RUB" {
		order, err := parser.GrinexBidUSDTRub()
		if err != nil {
			logger.Log.Errorf("failed to fetch: %v", err)
			return -1, err
		}
		gain := order.Price * amount
		fee, err := GetFee(domain.GrinexSource, domain.Usdtrub)
		if err != nil {
			logger.Log.Errorf("failed to get fee: %v", err)
			return -1, err
		}
		gain = gain * (1 - fee)

		return gain, nil
	} else if to + "/" + from == "USDT/RUB" {
		order, err := parser.GrinexAskUSDTRub()
		if err != nil {
			logger.Log.Errorf("failed to fetch: %v", err)
			return -1, err
		}
		gain := amount / order.Price
		fee, err := GetFee(domain.GrinexSource, domain.Usdtrub)
		if err != nil {
			logger.Log.Errorf("failed to get fee: %v", err)
			return -1, err
		}
		gain = gain * (1 - fee)

		return gain, nil
	} else if from + "/" + to == "A7A5/RUB" {
		order, err := parser.GrinexBidRubA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch: %v", err)
			return -1, err
		}
		gain := order.Price * amount
		fee, err := GetFee(domain.GrinexSource, domain.A7a5rub)
		if err != nil {
			logger.Log.Errorf("failed to get fee: %v", err)
			return -1, err
		}
		gain = gain * (1 - fee)

		return gain, nil
	} else if to + "/" + from == "A7A5/RUB" {
		order, err := parser.GrinexAskRubA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch: %v", err)
			return -1, err
		}
		gain := amount / order.Price
		fee, err := GetFee(domain.GrinexSource, domain.A7a5rub)
		if err != nil {
			logger.Log.Errorf("failed to get fee: %v", err)
			return -1, err
		}
		gain = gain * (1 - fee)

		return gain, nil
	} else if from + "/" + to == "USDT/A7A5" {
		order, err := parser.GrinexBidUSDTA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch: %v", err)
			return -1, err
		}
		gain := order.Price * amount
		fee, err := GetFee(domain.GrinexSource, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to get fee: %v", err)
			return -1, err
		}
		gain = gain * (1 - fee)

		return gain, nil
	} else if to + "/" + from == "USDT/A7A5" {
		order, err := parser.GrinexAskUSDTA7A5()
		if err != nil {
			logger.Log.Errorf("failed to fetch: %v", err)
			return -1, err
		}
		gain := amount / order.Price
		fee, err := GetFee(domain.GrinexSource, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to get fee: %v", err)
			return -1, err
		}
		gain = gain * (1 - fee)

		return gain, nil
	} else {
		err := errors.New("failed to convert")
		logger.Log.Errorf("failed to convert: %v", err)
		return -1, err
	}
}