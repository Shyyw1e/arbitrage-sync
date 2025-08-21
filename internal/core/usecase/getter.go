package usecase

import (
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/cache"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/parser"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

func getParsedData() (
	[]*domain.Order, []*domain.Order,
	[]*domain.Order, []*domain.Order,
//	[]*domain.Order, []*domain.Order,
	){
	rapiraRed, err := cache.GlobalOrderCache.GetOrFetch(
		cache.OrderCacheKey{
		Source: domain.RapiraSource,
		Pair:   domain.Usdtrub,
		Side:   domain.SideBuy,
		}, func() ([]*domain.Order, error) {
			a, err := parser.FetchRapiraAsk()
			return a, err
		},
	)
	if err != nil {
		logger.Log.Errorf("failed to fetch Rapira ask table: %v", err)
	}
	logger.Log.Infof("Got rapira: %v", len(rapiraRed))
	rapiraGreen, err := cache.GlobalOrderCache.GetOrFetch(
		cache.OrderCacheKey{
		Source: domain.RapiraSource,
		Pair:   domain.Usdtrub,
		Side:   domain.SideSell,
		}, func() ([]*domain.Order, error) {
			a, err := parser.FetchRapiraBid()
			return a, err
		},
	)
	if err != nil {
		logger.Log.Errorf("failed to fetch Rapira bid table: %v", err)
	}
	logger.Log.Infof("Got rapira: %v", len(rapiraGreen))
	// GrinexUSDTRUBRed, err := cache.GlobalOrderCache.GetOrFetch(
	// 	cache.OrderCacheKey{
	// 	Source: domain.GrinexUSDTRUBSource,
	// 	Pair:   domain.Usdtrub,
	// 	Side:   domain.SideBuy,
	// 	}, func() ([]*domain.Order, error) {
	// 		a, err := parser.FetchGrinexAskUSDTRub()
	// 		return a, err
	// 	},
	// )
	// if err != nil {
	// 	logger.Log.Errorf("failed to fetch Grinex USDT/RUB ask table: %v", err)
	// 	GrinexUSDTRUBRed = []*domain.Order{}
	// }
	// logger.Log.Infof("Got Grinex Ask USDT/RUB: %v", len(GrinexUSDTRUBRed))
	// GrinexUSDTRUBGreen, err :=cache.GlobalOrderCache.GetOrFetch(
	// 	cache.OrderCacheKey{
	// 	Source: domain.GrinexUSDTRUBSource,
	// 	Pair:   domain.Usdtrub,
	// 	Side:   domain.SideSell,
	// 	}, func() ([]*domain.Order, error) {
	// 		a, err := parser.FetchGrinexBidUSDTRub()
	// 		return a, err
	// 	},
	// )
	// if err != nil {
	// 	logger.Log.Errorf("failed to fetch Grinex USDT/RUB bid table: %v", err)
	// 	GrinexUSDTRUBGreen = []*domain.Order{}
	// }
	// logger.Log.Infof("Got Grinex Bid USDT/RUB: %v", len(GrinexUSDTRUBGreen))
	GrinexUSDTA7A5Red, err := cache.GlobalOrderCache.GetOrFetch(
		cache.OrderCacheKey{
		Source: domain.GrinexUSDTA7A5Source,
		Pair:   domain.Usdta7a5,
		Side:   domain.SideBuy,
		}, func() ([]*domain.Order, error) {
			a, err := parser.FetchGrinexAskUSDTA7A5()
			return a, err
		},
	)
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/A7A5 ask table: %v", err)
		GrinexUSDTA7A5Red = []*domain.Order{}
	}
	logger.Log.Infof("Got Grinex Ask USDT/A7A5: %v", len(GrinexUSDTA7A5Red))
	GrinexUSDTA7A5Green, err := cache.GlobalOrderCache.GetOrFetch(
		cache.OrderCacheKey{
		Source: domain.GrinexUSDTA7A5Source,
		Pair:   domain.Usdta7a5,
		Side:   domain.SideSell,
		}, func() ([]*domain.Order, error) {
			a, err := parser.FetchGrinexBidUSDTA7A5()
			return a, err
		},
	)
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/A7A5 bid table: %v", err)
		GrinexUSDTA7A5Green = []*domain.Order{}
	}
	logger.Log.Infof("Got Grinex Bid USDT/A7A5: %v", len(GrinexUSDTA7A5Green))


	return rapiraRed, rapiraGreen, GrinexUSDTA7A5Red, GrinexUSDTA7A5Green
}