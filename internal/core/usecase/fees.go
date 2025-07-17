package usecase

import (
	"errors"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

const (
	RapiraMaker = 0
	RapiraTaker = 0.001
)

var GrinexFee = map[domain.Pair]float64{
	"USDT/A7A5": 0.0005,
	"USD/A7A5":  0.0005,
	"USDT/RUB":  0.001,
	"A7A5/RUB":  0.0005,
}

func GetFee(source domain.Source, pair domain.Pair) (float64, error) {
	switch source {
	case domain.GrinexSource:
		if fee, ok := GrinexFee[pair]; ok {
			return fee, nil
		} else {
			logger.Log.Error("no such pair")
			return -1, errors.New("no such pair")
		}
		
	case domain.RapiraSource:
		return RapiraMaker, nil
		
	default:
		logger.Log.Error("invalid source")
		return -1, errors.New("invalid source")
	}
}