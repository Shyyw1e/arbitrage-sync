package domain

import "time"

type OrderSide string
type Pair string
type Direction string
type Source string

const(
	SideBuy	 OrderSide 	= "buy"			//красный стакан
	SideSell OrderSide 	= "sell"		//зеленый стакан
	
	RapiraSource Source = "rapira"
	GrinexUSDTRUBSource Source = "grinex USDT/RUB"
	GrinexUSDTA7A5Source Source = "grinex USDT/A7A5"

	Usdta7a5	 Pair = "USDT/A7A5" 
	Usdtrub		 Pair = "USDT/RUB"
	A7a5rub		 Pair = "A7A5/RUB"

	USDT		 Direction = "USDT"
	A7A5		 Direction = "A7A5"
	RUB		 	 Direction = "RUB"
)

type Order struct {
	Price 	float64
	Amount 	float64
	Sum		float64
	Side 	OrderSide
	Source 	Source		// rialto
	Pair   	Pair
}

type Opportunity struct {
    BuyExchange   Source  
	SellExchange  Source     
	BuyPrice      float64
	SellPrice     float64
	BuyPair       Pair       
	SellPair      Pair       
	BuyAmount     float64
	ProfitMargin  float64    
	SuggestedBid  float64    
	CreatedAt     time.Time
}