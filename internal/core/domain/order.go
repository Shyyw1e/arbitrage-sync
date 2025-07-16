package domain

import (

)

type OrderSide string
type Pair string
type Direction string

const(
	SideBuy	 OrderSide 	= "buy"			//красный стакан
	SideSell OrderSide 	= "sell"		//зеленый стакан
	
	RapiraSource string = "rapira"
	GrinexSource string = "grinex"

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
	Source 	string		// rialto
	Pair   	Pair
}

