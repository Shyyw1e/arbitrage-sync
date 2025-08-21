package usecase

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

var (
	recentMu sync.Mutex
	RecentHashes = make(map[string]time.Time)
)

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
			logger.Log.Warn("Sum overflowed")
			return opps, nil
		}

			// Effective prices
		effectiveAsk := ask.Price / (1 + feeAsk)
		effectiveBid := bid.Price * (1 + feeBid)
		profit := math.Round((effectiveAsk - effectiveBid)*100) / 100

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

func DetectFactArbitrage(
	ask *domain.Order,
	bid *domain.Order,
	minDiff, maxSum, feeAsk, feeBid float64,
	sourceAsk, sourceBid domain.Source,
	pair domain.Pair,
) ([]*domain.Opportunity, error) {
	opps := []*domain.Opportunity{}
	effectiveAsk := ask.Price / (1 + feeAsk)
	effectiveBid := bid.Price * (1 + feeBid)
	profit := math.Round((effectiveAsk - effectiveBid)*100) / 100
	
	logger.Log.Infof("Checking ask %.2f (eff %.4f) vs bid %.2f (eff %.4f) = profit %.4f",
		ask.Price, effectiveAsk, bid.Price, effectiveBid, profit)

	if profit >= minDiff {
		opportunity := &domain.Opportunity{
			BuyExchange:   	sourceBid,
			SellExchange:  	sourceAsk,
			ProfitMargin:  	profit,
			BuyPrice: 	   	math.Round(bid.Price*100) / 100,
			SellPrice: 	   	math.Round(ask.Price*100) / 100,
			BuyAmount:     	bid.Amount,
			SuggestedBid:  	bid.Price + 0.01,
			CreatedAt:     	time.Now(),
		}
		logger.Log.Infof(
			"Found fact arbitrage: Buy %s @ %.2f, Sell %s @ %.2f, Profit: %.4f",
			sourceBid, bid.Price, sourceAsk, ask.Price, profit,
		)
		opps = append(opps, opportunity)
	}
	if len(opps) == 0 {
		logger.Log.Info("No factic arbitrage situation!")
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
			logger.Log.Warn("Sum overflowed")
			return opps, nil
		}

			// Effective prices
		effectiveAsk := ask.Price / (1 + feeAsk)
		effectiveBid := bid.Price * (1 + feeBid)
		profit := math.Round((effectiveAsk - effectiveBid)*100) / 100

		logger.Log.Infof("Checking ask %.2f (eff %.4f) vs bid %.2f (eff %.4f) = profit %.4f",
			ask.Price, effectiveAsk, bid.Price, effectiveBid, profit)

		if profit >= minDiff {
			opportunity := &domain.Opportunity{
				BuyExchange:   	sourceBid,
				SellExchange:  	sourceAsk,
				ProfitMargin:  	profit,
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


func DetectAS(minDiff, maxSum float64, chatID int64) ([]*domain.Opportunity, []*domain.Opportunity, error) {
	CleanUpRecentHashes()
	opportunities := []*domain.Opportunity{}
	potential := []*domain.Opportunity{}

	rapiraRed, rapiraGreen,
		GrinexUSDTA7A5Red, GrinexUSDTA7A5Green := getParsedData()

	if len(rapiraRed) != 0 && len(rapiraGreen) != 0 {
		opportunityRapiraRG, err := DetectPairArbitrage(rapiraRed[1:], rapiraGreen[1:], minDiff, maxSum, 0.0, 0.0, domain.RapiraSource, domain.RapiraSource, domain.Usdtrub)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}

		opportunities = appendLastNonEmpty(opportunities, opportunityRapiraRG)
	}

	// if len(GrinexUSDTRUBRed) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	opportunityGrinexUSDTRUB, err := DetectPairArbitrage(GrinexUSDTRUBRed[1:], GrinexUSDTRUBGreen[1:], minDiff, maxSum, 0.001, 0.001, domain.GrinexUSDTRUBSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	opportunities = appendLastNonEmpty(opportunities, opportunityGrinexUSDTRUB)
	// }
	
	if len(GrinexUSDTA7A5Red) != 0 && len(GrinexUSDTA7A5Green) != 0 {
		opportunityGrinexUSDTA7A5, err := DetectPairArbitrage(GrinexUSDTA7A5Red[1:], GrinexUSDTA7A5Green[1:], minDiff, maxSum, 0.0005, 0.0005, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}

		opportunities = appendLastNonEmpty(opportunities, opportunityGrinexUSDTA7A5)
	}

	// if len(rapiraRed) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	opportunityRapiraAskGrinexUSDTRUBBid, err := DetectPairArbitrage(rapiraRed[1:], GrinexUSDTRUBGreen[1:], minDiff, maxSum, 0.0, 0.001, domain.RapiraSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	opportunities = appendLastNonEmpty(opportunities, opportunityRapiraAskGrinexUSDTRUBBid)
	// }

	// if len(GrinexUSDTRUBRed) != 0 && len(rapiraGreen) != 0 {
	// 	opportunityGrinexUSDTRUBAskRapiraBid, err := DetectPairArbitrage(GrinexUSDTRUBRed[1:], rapiraGreen[1:], minDiff, maxSum, 0.001, 0.0, domain.GrinexUSDTRUBSource, domain.RapiraSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	opportunities = appendLastNonEmpty(opportunities, opportunityGrinexUSDTRUBAskRapiraBid)
	// }

	if len(GrinexUSDTA7A5Red) != 0 && len(rapiraGreen) != 0 {
		opportunityGrinexUSDTA7A5AskRapiraBid, err := DetectPairArbitrage(GrinexUSDTA7A5Red[1:], rapiraGreen[1:], minDiff, maxSum, 0.0005, 0.0, domain.GrinexUSDTA7A5Source, domain.RapiraSource, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}

		opportunities = appendLastNonEmpty(opportunities, opportunityGrinexUSDTA7A5AskRapiraBid)
	}

	if len(rapiraRed) != 0 && len(GrinexUSDTA7A5Green) != 0 {
		opportunityRapiraAskGrinexUSDTA7A5Bid, err := DetectPairArbitrage(rapiraRed[1:], GrinexUSDTA7A5Green[1:], minDiff, maxSum, 0.0, 0.0005, domain.RapiraSource, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}
		
		opportunities = appendLastNonEmpty(opportunities, opportunityRapiraAskGrinexUSDTA7A5Bid)
	}

	// if len(GrinexUSDTRUBRed) != 0 && len(GrinexUSDTA7A5Green) != 0 {
	// 	opportunityGrinexUSDTRUBAskGrinexUSDTA7A5Bid, err := DetectPairArbitrage(GrinexUSDTRUBRed[1:], GrinexUSDTA7A5Green[1:], minDiff, maxSum, 0.001, 0.0005, domain.GrinexUSDTRUBSource, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}
		
	// 	opportunities = appendLastNonEmpty(opportunities, opportunityGrinexUSDTRUBAskGrinexUSDTA7A5Bid)
	// }

	// if len(GrinexUSDTA7A5Red) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	opportunityGrinexUSDTA7A5AskGrinexUSDTRUBBid, err := DetectPairArbitrage(GrinexUSDTA7A5Red[1:], GrinexUSDTRUBGreen[1:], minDiff, maxSum, 0.0005, 0.001, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTRUBSource, domain.Usdta7a5)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	opportunities = appendLastNonEmpty(opportunities, opportunityGrinexUSDTA7A5AskGrinexUSDTRUBBid)
	// }


	if len(rapiraRed) != 0 && len(rapiraGreen) != 0 {
		oppPotentialRap, err := DetectPairPotential(rapiraRed[1:], rapiraGreen[1:], minDiff, maxSum, 0.0, 0.0, domain.RapiraSource, domain.RapiraSource, domain.Usdtrub)
		if err != nil {
			logger.Log.Errorf("failed to detect potential situation: %v", err)
		}

		potential = appendLastNonEmpty(potential, oppPotentialRap)
	}

	// if len(GrinexUSDTRUBRed) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	oppPotentialGrUSDTRUB, err := DetectPairPotential(GrinexUSDTRUBRed[1:], GrinexUSDTRUBGreen[1:], minDiff, maxSum, 0.001, 0.001, domain.GrinexUSDTRUBSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect potential situation: %v", err)
	// 	}

	// 	potential = appendLastNonEmpty(potential, oppPotentialGrUSDTRUB)
	// }

	if len(GrinexUSDTA7A5Red) != 0 && len(GrinexUSDTA7A5Green) != 0 {
		oppPotentialGrUsdtA7a5, err := DetectPairPotential(GrinexUSDTA7A5Red[1:], GrinexUSDTA7A5Green[1:], minDiff, maxSum, 0.0005, 0.0005, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to detect potential situation: %v", err)
		}

		potential = appendLastNonEmpty(potential, oppPotentialGrUsdtA7a5)
	}

	// if len(rapiraRed) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	oppPotentialRapGRusdtRub, err := DetectPairPotential(rapiraRed[1:], GrinexUSDTRUBGreen[1:], minDiff, maxSum, 0.0, 0.001, domain.RapiraSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect potential situation: %v", err)
	// 	}

	// 	potential = appendLastNonEmpty(potential, oppPotentialRapGRusdtRub)
	// }

	// if len(GrinexUSDTRUBRed) != 0 && len(rapiraGreen) != 0 {
	// 	oppPotentialGrusdtRubRap, err := DetectPairPotential(GrinexUSDTRUBRed[1:], rapiraGreen[1:], minDiff, maxSum, 0.001, 0.0, domain.GrinexUSDTRUBSource, domain.RapiraSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect potential situation: %v", err)
	// 	}

	// 	potential = appendLastNonEmpty(potential, oppPotentialGrusdtRubRap)
	// }

	if len(rapiraRed) != 0 && len(GrinexUSDTA7A5Green) != 0 {
		oppPotentialRapGrUsdtA7A5, err := DetectPairPotential(rapiraRed[1:], GrinexUSDTA7A5Green[1:], minDiff, maxSum, 0.0, 0.0005, domain.RapiraSource, domain.GrinexUSDTA7A5Source, domain.Usdtrub)
		if err != nil {
			logger.Log.Errorf("failed to detect potential situation: %v", err)
		}

		potential = appendLastNonEmpty(potential, oppPotentialRapGrUsdtA7A5)
	}

	if len(GrinexUSDTA7A5Red) != 0 && len(rapiraGreen) != 0 {
		oppPotentialGrUsdtA7a5Rap, err := DetectPairPotential(GrinexUSDTA7A5Red[1:], rapiraGreen[1:], minDiff, maxSum, 0.0005, 0.0, domain.GrinexUSDTA7A5Source, domain.RapiraSource, domain.Usdtrub)
		if err != nil {
			logger.Log.Errorf("failed to detect potential situation: %v", err)
		}

		potential = appendLastNonEmpty(potential, oppPotentialGrUsdtA7a5Rap)
	}

	// if len(GrinexUSDTRUBRed) != 0 && len(GrinexUSDTA7A5Green) != 0 {
	// 	oppPotentialGrinexUsdtRubGrA7a5, err := DetectPairPotential(GrinexUSDTRUBRed[1:], GrinexUSDTA7A5Green[1:], minDiff, maxSum, 0.001, 0.0005, domain.GrinexUSDTRUBSource, domain.GrinexUSDTA7A5Source, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect potential situation: %v", err)
	// 	}

	// 	potential = appendLastNonEmpty(potential, oppPotentialGrinexUsdtRubGrA7a5)
	// }

	// if len(GrinexUSDTA7A5Red) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	oppPotentialGrinexA7a5GrUsdtRub, err := DetectPairPotential(GrinexUSDTA7A5Red[1:], GrinexUSDTRUBGreen[1:], minDiff, maxSum, 0.0005, 0.001, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect potential situation: %v", err)
	// 	}

	// 	potential = appendLastNonEmpty(potential, oppPotentialGrinexA7a5GrUsdtRub)
	// }

	if len(opportunities) == 0 && len(potential) == 0 {
		logger.Log.Warn("empty orderbook; skip tick")
		return nil, nil, nil
	}


	if len(potential) > 0 || len(opportunities) > 0{
		logger.Log.Info("Arbitrage situation detected:\n")
		for _, el := range opportunities {
			logger.Log.Infof("Buy exchange: %v\tSell exchange: %v\nBuy price: %v\tSell price: %v\nBuy amount: %v\tProfit margin: %v\n Full profit %v\n",
				el.BuyExchange, el.SellExchange, el.BuyPrice, el.SellPrice, el.BuyAmount, el.ProfitMargin, el.BuyAmount * (1 + el.ProfitMargin))
		}	
		return opportunities, potential, nil
	} else {
		logger.Log.Info("Arbitrage situation not detected")
		return nil, nil, nil
	}
}

func DetectFact(minDiff, maxSum float64, chatID int64) ([]*domain.Opportunity, error) {
	facticOpp := []*domain.Opportunity{}
	
	rapiraRed, rapiraGreen,
	GrinexUSDTA7A5Red, GrinexUSDTA7A5Green := getParsedData()
	logger.Log.Info("Getting facts")

	if len(rapiraRed) != 0 && len(rapiraGreen) != 0 {
		factRapiraRG, err := DetectFactArbitrage(rapiraRed[0], rapiraGreen[0], minDiff, maxSum, 0.0, 0.0, domain.RapiraSource, domain.RapiraSource, domain.Usdtrub)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}

		logger.Log.Info("Appending factic")
		facticOpp = appendLastNonEmpty(facticOpp, factRapiraRG)
	}

	// if len(GrinexUSDTRUBRed) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	factGrinexUSDTRUB, err := DetectFactArbitrage(GrinexUSDTRUBRed[0], GrinexUSDTRUBGreen[0], minDiff, maxSum, 0.001, 0.001, domain.GrinexUSDTRUBSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	logger.Log.Info("Appending factic")
	// 	facticOpp = appendLastNonEmpty(facticOpp, factGrinexUSDTRUB)
	// }

	if len(GrinexUSDTA7A5Red) != 0 && len(GrinexUSDTA7A5Green) != 0 {
		factGrinexUSDTA7A5, err := DetectFactArbitrage(GrinexUSDTA7A5Red[0], GrinexUSDTA7A5Green[0], minDiff, maxSum, 0.0005, 0.0005, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}

		facticOpp = appendLastNonEmpty(facticOpp, factGrinexUSDTA7A5)
	}

	// if len(rapiraRed) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	factRapiraAskGrinexUSDTRUBBid, err := DetectFactArbitrage(rapiraRed[0], GrinexUSDTRUBGreen[0], minDiff, maxSum, 0.0, 0.001, domain.RapiraSource, domain.GrinexUSDTRUBSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	logger.Log.Info("Appending factic")
	// 	facticOpp = appendLastNonEmpty(facticOpp, factRapiraAskGrinexUSDTRUBBid)
	// }

	// if len(GrinexUSDTRUBRed) != 0 && len(rapiraRed) != 0 {
	// 	factGrinexUSDTRUBAskRapiraBid, err := DetectFactArbitrage(GrinexUSDTRUBRed[0], rapiraGreen[0], minDiff, maxSum, 0.001, 0.0, domain.GrinexUSDTRUBSource, domain.RapiraSource, domain.Usdtrub)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	logger.Log.Info("Appending factic")
	// 	facticOpp = appendLastNonEmpty(facticOpp, factGrinexUSDTRUBAskRapiraBid)
	// }

	if len(GrinexUSDTA7A5Red) != 0 && len(rapiraGreen) != 0 {
	factGrinexUSDTA7A5AskRapiraBid, err := DetectFactArbitrage(GrinexUSDTA7A5Red[0], rapiraGreen[0], minDiff, maxSum, 0.0005, 0.0, domain.GrinexUSDTA7A5Source, domain.RapiraSource, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}

		logger.Log.Info("Appending factic")
		facticOpp = appendLastNonEmpty(facticOpp, factGrinexUSDTA7A5AskRapiraBid)
	}

	if len(rapiraRed) != 0 && len(GrinexUSDTA7A5Green) != 0 {
	factRapiraAskGrinexUSDTA7A5Bid, err := DetectFactArbitrage(rapiraRed[0], GrinexUSDTA7A5Green[0], minDiff, maxSum, 0.0, 0.0005, domain.RapiraSource, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
		if err != nil {
			logger.Log.Errorf("failed to detect AS: %v", err)
		}
		
		logger.Log.Info("Appending factic")
		facticOpp = appendLastNonEmpty(facticOpp, factRapiraAskGrinexUSDTA7A5Bid)
	}

	// if len(GrinexUSDTRUBRed) != 0 && len(GrinexUSDTA7A5Green) != 0 {
	// 	factGrinexUSDTRUBAskGrinexUSDTA7A5Bid, err := DetectFactArbitrage(GrinexUSDTRUBRed[0], GrinexUSDTA7A5Green[0], minDiff, maxSum, 0.001, 0.0005, domain.GrinexUSDTRUBSource, domain.GrinexUSDTA7A5Source, domain.Usdta7a5)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}
		
	// 	logger.Log.Info("Appending factic")
	// 	facticOpp = appendLastNonEmpty(facticOpp, factGrinexUSDTRUBAskGrinexUSDTA7A5Bid)
	// }

	// if len(GrinexUSDTA7A5Red) != 0 && len(GrinexUSDTRUBGreen) != 0 {
	// 	factGrinexUSDTA7A5AskGrinexUSDTRUBBid, err := DetectFactArbitrage(GrinexUSDTA7A5Red[0], GrinexUSDTRUBGreen[0], minDiff, maxSum, 0.0005, 0.001, domain.GrinexUSDTA7A5Source, domain.GrinexUSDTRUBSource, domain.Usdta7a5)
	// 	if err != nil {
	// 		logger.Log.Errorf("failed to detect AS: %v", err)
	// 	}

	// 	logger.Log.Info("Appending factic")
	// 	facticOpp = appendLastNonEmpty(facticOpp, factGrinexUSDTA7A5AskGrinexUSDTRUBBid)
	// }	

	if len(facticOpp) == 0{
		logger.Log.Warn("empty orderbook; skip tick")
		return nil, nil
	}

	if len(facticOpp) > 0 {
		logger.Log.Info("Arbitrage situation detected:\n")
		for _, el := range facticOpp {
			logger.Log.Infof("Buy exchange: %v\tSell exchange: %v\nBuy price: %v\tSell price: %v\nBuy amount: %v\tProfit margin: %v\n Full profit %v\n",
				el.BuyExchange, el.SellExchange, el.BuyPrice, el.SellPrice, el.BuyAmount, el.ProfitMargin, el.BuyAmount * (1 + el.ProfitMargin))
		}	
		return facticOpp, nil
	} else {
		logger.Log.Info("Arbitrage situation not detected")
		return nil, nil
	}

}

func HashOpportunity(op *domain.Opportunity, chatID int64) string {
	return fmt.Sprintf("%d-%s-%s-%f-%f-%f", chatID, op.BuyExchange, op.SellExchange, op.BuyPrice, op.SellPrice, op.BuyAmount)
}

func CleanUpRecentHashes() {
	now := time.Now()
	recentMu.Lock()
	for hash, t := range RecentHashes {
		if now.Sub(t) > time.Hour {
			delete(RecentHashes, hash)
		}
	}
	recentMu.Unlock()
}

func markAsSeen(hash string) {
	recentMu.Lock()
	RecentHashes[hash] = time.Now()
	recentMu.Unlock()
}

func appendLastNonEmpty(dst []*domain.Opportunity, src []*domain.Opportunity) []*domain.Opportunity {
	if len(src) > 0 {
		return append(dst, src[len(src)-1])
	}
	return dst
}
