package parser

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/chromedp/chromedp"
)

func FetchGrinexAskUSDTRub() ([]*domain.Order, error) {
	return fetchGrinexOrders("https://grinex.io/trading/usdtrub", domain.Usdtrub, domain.SideSell, "ask_orders_panel", "usdtrub")
}

func FetchGrinexBidUSDTRub() ([]*domain.Order, error) {
	return fetchGrinexOrders("https://grinex.io/trading/usdtrub", domain.Usdtrub, domain.SideBuy, "bid_orders_panel", "usdtrub")
}

func FetchGrinexAskUSDTA7A5() ([]*domain.Order, error) {
	return fetchGrinexOrders("https://grinex.io/trading/usdta7a5", domain.Usdta7a5, domain.SideSell, "ask_orders_panel", "usdta7a5")
}

func FetchGrinexBidUSDTA7A5() ([]*domain.Order, error) {
	return fetchGrinexOrders("https://grinex.io/trading/usdta7a5", domain.Usdta7a5, domain.SideBuy, "bid_orders_panel", "usdta7a7")
}

func fetchGrinexOrders(url string, pair domain.Pair, side domain.OrderSide, panelClass string, marketTab string) ([]*domain.Order, error) {
	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(), chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	selector := fmt.Sprintf(`div#order_book_holder[data-market="%s_tab"] div.%s table`, marketTab, panelClass)

	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(selector+" tbody tr", chromedp.ByQuery),
		chromedp.OuterHTML(selector, &html, chromedp.ByQuery),
	)
	if err != nil {
		logger.Log.Errorf("failed to load Grinex orders: %v", err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logger.Log.Errorf("failed to parse HTML: %v", err)
		return nil, err
	}

	var orders []*domain.Order
	rows := doc.Find("tbody tr")
	for i := 0; i < 5 && i < rows.Length(); i++ {
		row := rows.Eq(i)
		cols := row.Find("td")
		if cols.Length() < 3 {
			logger.Log.Errorf("invalid length: %v", cols.Length())
			continue
		}

		priceText := cols.Eq(0).Text()
		amountText := cols.Eq(1).Text()
		sumText := cols.Eq(2).Text()

		price, err1 := parseGrinexNumber(priceText)
		amount, err2 := parseGrinexNumber(amountText)
		sum, err3 := parseGrinexNumber(sumText)
		if err1 != nil || err2 != nil || err3 != nil {
			logger.Log.Errorf("failed to convert strings %v\t%v\t%v", err1, err2, err3)
			continue
		}

		var source domain.Source
		switch pair {
		case domain.Usdta7a5:
			source = domain.GrinexUSDTA7A5Source
		case domain.Usdtrub:
			source = domain.GrinexUSDTRUBSource
		}

		orders = append(orders, &domain.Order{
			Price:  price,
			Amount: amount,
			Sum:    sum,
			Side:   side,
			Source: source,
			Pair:   pair,
		})
	}

	if len(orders) == 0 {
		return nil, errors.New("no orders parsed from Grinex")
	}

	logger.Log.Infof("Parsed orders got: %v", len(orders))

	return orders, nil
}

func parseGrinexNumber(s string) (float64, error) {
	s = strings.ReplaceAll(s, "\u00a0", "") // nbsp
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	return strconv.ParseFloat(s, 64)
}