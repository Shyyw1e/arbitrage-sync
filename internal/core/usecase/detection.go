package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
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
) (*domain.Opportunity, bool, error) {
	if len(asks) == 0 || len(bids) == 0 {
		err := errors.New("empty orderbook")
		logger.Log.Errorf("failed to detect AS: %v", err)
		return nil, false, err
	}

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
				BuyExchange:   sourceBid,
				SellExchange:  sourceAsk,
				BuyPrice:      bid.Price,
				SellPrice:     ask.Price,
				BuyPair:       bid.Pair,
				SellPair:      ask.Pair,
				BuyAmount:     bid.Amount,
				ProfitMargin:  profit,
				SuggestedBid:  bid.Price + 0.01,
				CreatedAt:     time.Now(),
			}
			logger.Log.Infof(
				"Found arbitrage: Buy %s @ %.2f, Sell %s @ %.2f, Profit: %.4f",
				sourceBid, bid.Price, sourceAsk, ask.Price, profit,
			)
			return opportunity, true, nil
		}
	}
	

	return nil, false, nil
}


func DetectAS(minDiff, maxSum float64) ([]*domain.Opportunity, error) {
	CleanUpRecentHashes()
	opportunities := make([]*domain.Opportunity, 0)

	rapiraRed, err := parser.FetchRapiraAsk()
	if err != nil {
		logger.Log.Errorf("failed to fetch Rapira ask table: %v", err)
	}
	logger.Log.Infof("Got rapira: %v\trapiraRed[0].Price: %v", len(rapiraRed), rapiraRed[0].Price)
	rapiraGreen, err := parser.FetchRapiraBid()
	if err != nil {
		logger.Log.Errorf("failed to fetch Rapira bid table: %v", err)
	}
	logger.Log.Infof("Got rapira: %v\trapiraGreen[0].Price: %v", len(rapiraGreen), rapiraGreen[0].Price)
	GrinexUSDTRUBRed, err := parser.FetchGrinexAskUSDTRub()
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/RUB ask table: %v", err)
	}
	GrinexUSDTRUBGreen, err := parser.FetchGrinexBidUSDTRub()
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/RUB bid table: %v", err)
	}
	GrinexUSDTA7A5Red, err := parser.FetchGrinexAskUSDTA7A5()
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/A7A5 ask table: %v", err)
	}
	GrinexUSDTA7A5Green, err := parser.FetchGrinexBidUSDTA7A5()
	if err != nil {
		logger.Log.Errorf("failed to fetch Grinex USDT/A7A5 bid table: %v", err)
	}

	logger.Log.Info("Detecting Rapira AS")
	opportunityRapiraRG, ok, err := DetectPairArbitrage(rapiraRed, rapiraGreen, minDiff, maxSum, 0.0, 0.0, domain.RapiraSource, domain.RapiraSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
		logger.Log.Infof("Opportunity:\t%v\t%v\t%v\t%v",opportunityRapiraRG.BuyPrice, opportunityRapiraRG.SellPrice, opportunityRapiraRG.BuyExchange, opportunityRapiraRG.SellExchange)
	}
	if ok {
		opportunities = append(opportunities, opportunityRapiraRG)
	}
	
	opportunityGrinexUSDTRUB, ok, err := DetectPairArbitrage(GrinexUSDTRUBRed, GrinexUSDTRUBGreen, minDiff, maxSum, 0.001, 0.001, domain.GrinexSource, domain.GrinexSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityGrinexUSDTRUB)
	}
	
	opportunityGrinexUSDTA7A5, ok, err := DetectPairArbitrage(GrinexUSDTA7A5Red, GrinexUSDTA7A5Green, minDiff, maxSum, 0.0005, 0.0005, domain.GrinexSource, domain.GrinexSource, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityGrinexUSDTA7A5)
	}

	opportunityRapiraAskGrinexUSDTRUBBid, ok, err := DetectPairArbitrage(rapiraRed, GrinexUSDTRUBGreen, minDiff, maxSum, 0.0, 0.001, domain.RapiraSource, domain.GrinexSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityRapiraAskGrinexUSDTRUBBid)
	}

	opportunityGrinexUSDTRUBAskRapiraBid, ok, err := DetectPairArbitrage(GrinexUSDTRUBRed, rapiraGreen, minDiff, maxSum, 0.001, 0.0, domain.GrinexSource, domain.RapiraSource, domain.Usdtrub)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityGrinexUSDTRUBAskRapiraBid)
	}

	opportunityGrinexUSDTA7A5AskRapiraBid, ok, err := DetectPairArbitrage(GrinexUSDTA7A5Red, rapiraGreen, minDiff, maxSum, 0.0005, 0.0, domain.GrinexSource, domain.RapiraSource, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityGrinexUSDTA7A5AskRapiraBid)
	}

	opportunityRapiraAskGrinexUSDTA7A5Bid, ok, err := DetectPairArbitrage(rapiraRed, GrinexUSDTA7A5Green, minDiff, maxSum, 0.0, 0.0005, domain.RapiraSource, domain.GrinexSource, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityRapiraAskGrinexUSDTA7A5Bid)
	}

	opportunityGrinexUSDTRUBAskGrinexUSDTA7A5Bid, ok, err := DetectPairArbitrage(GrinexUSDTRUBRed, GrinexUSDTA7A5Green, minDiff, maxSum, 0.001, 0.0005, domain.GrinexSource, domain.GrinexSource, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityGrinexUSDTRUBAskGrinexUSDTA7A5Bid)
	}

	opportunityGrinexUSDTA7A5AskGrinexUSDTRUBBid, ok, err := DetectPairArbitrage(GrinexUSDTA7A5Red, GrinexUSDTRUBGreen, minDiff, maxSum, 0.0005, 0.001, domain.GrinexSource, domain.GrinexSource, domain.Usdta7a5)
	if err != nil {
		logger.Log.Errorf("failed to detect AS: %v", err)
	}
	if ok {
		opportunities = append(opportunities, opportunityGrinexUSDTA7A5AskGrinexUSDTRUBBid)
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
	if len(cleanOps) > 0 {
		logger.Log.Info("Arbitrage situation detected:\n")
		for _, el := range cleanOps {
			logger.Log.Infof("Buy exchange: %v\tSell exchange: %v\nBuy price: %v\tSell price: %v\nBuy amount: %v\tProfit margin: %v\n Full profit %v\n",
				el.BuyExchange, el.SellExchange, el.BuyPrice, el.SellPrice, el.BuyAmount, el.ProfitMargin, el.BuyAmount * (1 + el.ProfitMargin))
		}	
		return cleanOps, nil
	} else {
		logger.Log.Info("Arbitrage situation not detected")
		return nil, nil
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