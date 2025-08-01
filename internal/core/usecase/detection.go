package usecase

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/cache"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/parser"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

var RecentHashes = make(map[string]time.Time)

func DetectPairArbitrage(
	asks []*domain.Order,
	bids []*domain.Order,
	minDiff, maxSum, feeAsk, feeBid float64,
	sourceAsk, sourceBid domain.Source,
	pair domain.Pair,
) ([]*domain.Opportunity, error) {
	if len(asks) == 0 || len(bids) == 0 {
		err := errors.New("empty orderbook")
		logger.Log.Errorf("failed to detect AS: %v", err)
		return nil, err
	}
	opps := []*domain.Opportunity{}
	ask := asks[0]

	var accumulatedBidAmount float64
	for _, bid := range bids {
		if sourceBid == domain.RapiraSource {
			accumulatedBidAmount = bid.Sum
		} else {
			accumulatedBidAmount += bid.Amount
		}
		if accumulatedBidAmount > maxSum {
			break
		}

			// Effective prices
		effectiveAsk := ask.Price / (1 + feeAsk)
		effectiveBid := bid.Price * (1 + feeBid)
		profit := effectiveAsk - effectiveBid

		logger.Log.Infof("Checking ask %.2f (eff %.4f) vs bid %.2f (eff %.4f) = profit %.4f",
			ask.Price, effectiveAsk, bid.Price, effectiveBid, profit)

		if profit >= minDiff {
			opportunity := &domain.Opportunity{
				BuyExchange:   	sourceBid,
				SellExchange:  	sourceAsk,
				ProfitMargin:  	math.Round((effectiveAsk - effectiveBid)*100) / 100,
				BuyPrice: 	   	math.Round(bid.Price*100) / 100,
				SellPrice: 	   	math.Round(ask.Price*100) / 100,
				BuyAmount:     	bid.Amount,
				SuggestedBid:  	bid.Price + 0.01,
				CreatedAt:     	time.Now(),
			}
			logger.Log.Infof(
				"Found arbitrage: Buy %s @ %.2f, Sell %s @ %.2f, Profit: %.4f",
				sourceBid, bid.Price, sourceAsk, ask.Price, profit,
			)
			opps = append(opps, opportunity)
		}
	}
	

	return opps, nil
}

func DetectPairPotential(
	asks []*domain.Order,
	bids []*domain.Order,
	minDiff, maxSum, feeAsk, feeBid float64,
	sourceAsk, sourceBid domain.Source,
	pair domain.Pair,
) ([]*domain.Opportunity, error) {
	if len(asks) == 0 || len(bids) == 0 {
		err := errors.New("empty orderbook")
		logger.Log.Errorf("failed to detect AS: %v", err)
		return nil, err
	}
	opps := []*domain.Opportunity{}
	bid := bids[0]

	var accumulatedBidAmount float64
	for _, ask := range asks {
		if sourceAsk == domain.RapiraSource {
			accumulatedBidAmount = ask.Sum
		} else {
			accumulatedBidAmount += ask.Amount
		}
		if accumulatedBidAmount > maxSum {
			break
		}

			// Effective prices
		effectiveAsk := ask.Price / (1 + feeAsk)
		effectiveBid := bid.Price * (1 + feeBid)
		profit := effectiveAsk - effectiveBid

		logger.Log.Infof("Checking ask %.2f (eff %.4f) vs bid %.2f (eff %.4f) = profit %.4f",
			ask.Price, effectiveAsk, bid.Price, effectiveBid, profit)

		if profit >= minDiff {
			opportunity := &domain.Opportunity{
				BuyExchange:   	sourceBid,
				SellExchange:  	sourceAsk,
				ProfitMargin:  	math.Round((effectiveAsk - effectiveBid)*100) / 100,
				BuyPrice: 	   	math.Round(bid.Price*100) / 100,
				SellPrice: 	   	math.Round(ask.Price*100) / 100,
				BuyAmount:     	bid.Amount,
				SuggestedBid:  	bid.Price + 0.01,
				CreatedAt:     	time.Now(),
			}
			logger.Log.Infof(
				"Found potential: Buy %s @ %.2f, Sell %s @ %.2f, Profit: %.4f",
				sourceBid, bid.Price, sourceAsk, ask.Price, profit,
			)
			opps = append(opps, opportunity)
		}
	}
	

	return opps, nil

}


func DetectAS(minDiff, maxSum float64) ([]*domain.Opportunity, []*domain.Opportunity, error) {
	CleanUpRecentHashes()
	opportunities := make([]*domain.Opportunity, 0)
	potential := make([]*domain.Opportunity, 0)

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
	logger.Log.Infof("Got rapira: %v\trapiraRed[0].Price: %v", len(rapiraRed), rapiraRed[0].Price)
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
	logger.Log.Infof("Got rapira: %v\trapiraGreen[0].Price: %v", len(rapiraGreen), rapiraGreen[0].Price)
	GrinexUSDTRUBRed, err := cache.GlobalOrderCache.GetOrFetch(
		cache.OrderCacheKey{
		Source: domain.GrinexUSDTRUBSource,
		Pair:   domain.Usdtrub,
		Side:   domain.SideBuy,
		}, func() ([]*domain.Order, error) {
			a, err := parser.FetchGrinexAskUSDTRub()
			return a, err
		},
	)
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/RUB ask table: %v", err)
		GrinexUSDTRUBRed = []*domain.Order{}
	}
	logger.Log.Infof("Got Grinex USDT/RUB: %v", len(GrinexUSDTRUBRed))
	GrinexUSDTRUBGreen, err :=cache.GlobalOrderCache.GetOrFetch(
		cache.OrderCacheKey{
		Source: domain.GrinexUSDTRUBSource,
		Pair:   domain.Usdtrub,
		Side:   domain.SideSell,
		}, func() ([]*domain.Order, error) {
			a, err := parser.FetchGrinexBidUSDTRub()
			return a, err
		},
	)
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/RUB bid table: %v", err)
		GrinexUSDTRUBGreen = []*domain.Order{}
	}
	logger.Log.Infof("Got Grinex USDT/RUB: %v", len(GrinexUSDTRUBGreen))
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

	logger.Log.Info("Detecting Rapira AS")
	opportunityRapiraRG, err := DetectPairArbitrage(rapiraRed, rapiraGreen, minDiff, maxSum, 0.0, 0.0, domain.RapiraSource, domain.RapiraSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}

	opportunities = append(opportunities, opportunityRapiraRG[len(opportunityRapiraRG) - 1])

	opportunityGrinexUSDTRUB, err := DetectPairArbitrage(GrinexUSDTRUBRed, GrinexUSDTRUBGreen, minDiff, maxSum, 0.001, 0.001, domain.GrinexUSDTRUBSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}

	opportunities = append(opportunities, opportunityGrinexUSDTRUB[len(opportunityGrinexUSDTRUB) - 1])
	
	opportunityGrinexUSDTA7A5, err := DetectPairArbitrage(GrinexUSDTA7A5Red, GrinexUSDTA7A5Green, minDiff, maxSum, 0.0005, 0.0005, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}

	opportunities = append(opportunities, opportunityGrinexUSDTA7A5[len(opportunityGrinexUSDTA7A5) - 1])

	opportunityRapiraAskGrinexUSDTRUBBid, err := DetectPairArbitrage(rapiraRed, GrinexUSDTRUBGreen, minDiff, maxSum, 0.0, 0.001, domain.RapiraSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}

	opportunities = append(opportunities, opportunityRapiraAskGrinexUSDTRUBBid[len(opportunityRapiraAskGrinexUSDTRUBBid) - 1])

	opportunityGrinexUSDTRUBAskRapiraBid, err := DetectPairArbitrage(GrinexUSDTRUBRed, rapiraGreen, minDiff, maxSum, 0.001, 0.0, domain.GrinexUSDTRUBSource, domain.RapiraSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}

	opportunities = append(opportunities, opportunityGrinexUSDTRUBAskRapiraBid[len(opportunityGrinexUSDTRUBAskRapiraBid) - 1])

	opportunityGrinexUSDTA7A5AskRapiraBid, err := DetectPairArbitrage(GrinexUSDTA7A5Red, rapiraGreen, minDiff, maxSum, 0.0005, 0.0, domain.GrinexUSDTA7A5Source, domain.RapiraSource, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}

	opportunities = append(opportunities, opportunityGrinexUSDTA7A5AskRapiraBid[len(opportunityGrinexUSDTA7A5AskRapiraBid) - 1])


	opportunityRapiraAskGrinexUSDTA7A5Bid, err := DetectPairArbitrage(rapiraRed, GrinexUSDTA7A5Green, minDiff, maxSum, 0.0, 0.0005, domain.RapiraSource, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	
	opportunities = append(opportunities, opportunityRapiraAskGrinexUSDTA7A5Bid[len(opportunityRapiraAskGrinexUSDTA7A5Bid) - 1])


	opportunityGrinexUSDTRUBAskGrinexUSDTA7A5Bid, err := DetectPairArbitrage(GrinexUSDTRUBRed, GrinexUSDTA7A5Green, minDiff, maxSum, 0.001, 0.0005, domain.GrinexUSDTRUBSource, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	
	opportunities = append(opportunities, opportunityGrinexUSDTRUBAskGrinexUSDTA7A5Bid[len(opportunityGrinexUSDTRUBAskGrinexUSDTA7A5Bid) - 1])

	opportunityGrinexUSDTA7A5AskGrinexUSDTRUBBid, err := DetectPairArbitrage(GrinexUSDTA7A5Red, GrinexUSDTRUBGreen, minDiff, maxSum, 0.0005, 0.001, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTRUBSource, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}

	opportunities = append(opportunities, opportunityGrinexUSDTA7A5AskGrinexUSDTRUBBid[len(opportunityGrinexUSDTA7A5AskGrinexUSDTRUBBid) - 1])

	oppPotentialRap, err := DetectPairPotential(rapiraRed, rapiraGreen, minDiff, maxSum, 0.0, 0.0, domain.RapiraSource, domain.RapiraSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialRap[len(oppPotentialRap) - 1])

	oppPotentialGrUSDTRUB, err := DetectPairPotential(GrinexUSDTRUBRed, GrinexUSDTRUBGreen, minDiff, maxSum, 0.001, 0.001, domain.GrinexUSDTRUBSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialGrUSDTRUB[len(oppPotentialGrUSDTRUB) - 1])
	
	oppPotentialGrUsdtA7a5, err := DetectPairPotential(GrinexUSDTA7A5Red, GrinexUSDTA7A5Green, minDiff, maxSum, 0.0005, 0.0005, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialGrUsdtA7a5[len(oppPotentialGrUsdtA7a5) - 1])

	oppPotentialRapGRusdtRub, err := DetectPairPotential(rapiraRed, GrinexUSDTRUBGreen, minDiff, maxSum, 0.0, 0.001, domain.RapiraSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialRapGRusdtRub[len(oppPotentialRapGRusdtRub) - 1])

	oppPotentialGrusdtRubRap, err := DetectPairPotential(GrinexUSDTRUBRed, rapiraGreen, minDiff, maxSum, 0.001, 0.0, domain.GrinexUSDTRUBSource, domain.RapiraSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialGrusdtRubRap[len(oppPotentialGrusdtRubRap) - 1])

	oppPotentialRapGrUsdtA7A5, err := DetectPairPotential(rapiraRed, GrinexUSDTA7A5Green, minDiff, maxSum, 0.0, 0.0005, domain.RapiraSource, domain.GrinexUSDTA7A5Source, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialRapGrUsdtA7A5[len(oppPotentialRapGrUsdtA7A5) - 1])

	oppPotentialGrUsdtA7a5Rap, err := DetectPairPotential(GrinexUSDTA7A5Red, rapiraGreen, minDiff, maxSum, 0.0005, 0.0, domain.GrinexUSDTA7A5Source, domain.RapiraSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialGrUsdtA7a5Rap[len(oppPotentialGrUsdtA7a5Rap) - 1])

	oppPotentialGrinexUsdtRubGrA7a5, err := DetectPairPotential(GrinexUSDTRUBRed, GrinexUSDTA7A5Green, minDiff, maxSum, 0.001, 0.0005, domain.GrinexUSDTRUBSource, domain.GrinexUSDTA7A5Source, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialGrinexUsdtRubGrA7a5[len(oppPotentialGrinexUsdtRubGrA7a5) - 1])

	oppPotentialGrinexA7a5GrUsdtRub, err := DetectPairPotential(GrinexUSDTA7A5Red, GrinexUSDTRUBGreen, minDiff, maxSum, 0.0005, 0.001, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect potential situation: %v", err)
	}

	potential = append(potential, oppPotentialGrinexA7a5GrUsdtRub[len(oppPotentialGrinexA7a5GrUsdtRub) - 1])


	cleanPots := make([]*domain.Opportunity, 0)
	for _, op := range potential {
		hash := HashOpportunity(op)
		if _, seen := RecentHashes[hash]; seen {
			continue
		}
		markAsSeen(hash)
		cleanPots = append(cleanPots, op)
	}
	
	cleanOps := make([]*domain.Opportunity, 0)
	for _, op := range opportunities {
		hash := HashOpportunity(op)
		if _, seen := RecentHashes[hash]; seen {
			continue
		}
		markAsSeen(hash)
		cleanOps = append(cleanOps, op)
	}
	if len(cleanOps) > 0 || len(cleanPots) > 0 {
		logger.Log.Info("Arbitrage situation detected:\n")
		for _, el := range cleanOps {
			logger.Log.Infof("Buy exchange: %v\tSell exchange: %v\nBuy price: %v\tSell price: %v\nBuy amount: %v\tProfit margin: %v\n Full profit %v\n",
				el.BuyExchange, el.SellExchange, el.BuyPrice, el.SellPrice, el.BuyAmount, el.ProfitMargin, el.BuyAmount * (1 + el.ProfitMargin))
		}	
		return cleanOps, cleanPots, nil
	} else {
		logger.Log.Info("Arbitrage situation not detected")
		return nil, nil, nil
	}
}

func HashOpportunity(op *domain.Opportunity) string {
	return fmt.Sprintf("%s-%s-%f-%f-%f", op.BuyExchange, op.SellExchange, op.BuyPrice, op.SellPrice, op.BuyAmount)
}

func CleanUpRecentHashes() {
	now := time.Now()
	for hash, t := range RecentHashes {
		if now.Sub(t) > time.Hour {
			delete(RecentHashes, hash)
		}
	}
}

func markAsSeen(hash string) {
	RecentHashes[hash] = time.Now()
}