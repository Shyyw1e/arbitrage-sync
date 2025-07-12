package domain

import (

)

type OrderSide string

const(
	SideBuy	 OrderSide 	= "buy"		//красный стакан
	SideSell OrderSide 	= "sell"		//зеленый стакан
	RapiraSource string = "rapira"
	GrinexSource string = "grinex" 
)

type Order struct {
	Price 	float64
	Amount 	float64
	Side 	OrderSide
	Source 	string		// rialto
}

