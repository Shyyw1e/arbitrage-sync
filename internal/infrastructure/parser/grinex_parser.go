package parser

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	"github.com/chromedp/chromedp"
)



// <div class="ask_orders_panel|bid_orders_panel"><table>...<tbody class="usdtrub_ask asks">...</tbody></table></div>

func FetchGrinexAskUSDTRub() ([]*domain.Order, error) {
	logger.Log.Info("Fetching Grinex USDT/RUB ask")
    return fetchGrinexOrdersUSDTRUB(
        "https://grinex.io/trading/usdtrub",
        domain.Usdtrub,
        domain.SideSell,
        "ask_orders_panel",
        "usdtrub",
    )
}

func FetchGrinexBidUSDTRub() ([]*domain.Order, error) {
	logger.Log.Info("Fetching Grinex USDT/RUB bid")
    return fetchGrinexOrdersUSDTRUB(
        "https://grinex.io/trading/usdtrub",
        domain.Usdtrub,
        domain.SideBuy,
        "bid_orders_panel",
        "usdtrub",
    )
}

func FetchGrinexAskUSDTA7A5() ([]*domain.Order, error) {
	logger.Log.Info("Fetching Grinex USDT/A7A5 ask")
    return fetchGrinexOrdersUSDTA7A5(
        "https://grinex.io/trading/usdta7a5",
        domain.Usdta7a5,
        domain.SideSell,
        "ask_orders_panel",
        "usdta7a5",
    )
}

func FetchGrinexBidUSDTA7A5() ([]*domain.Order, error) {
    logger.Log.Info("Fetching Grinex USDT/A7A5 bid")
	return fetchGrinexOrdersUSDTA7A5(
        "https://grinex.io/trading/usdta7a5",
        domain.Usdta7a5,
        domain.SideBuy,
        "bid_orders_panel",
        "usdta7a5",
    )
}


func fetchGrinexOrdersUSDTRUB(url string, pair domain.Pair, side domain.OrderSide, panelClass string, marketTab string) ([]*domain.Order, error) {
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
        return nil, fmt.Errorf("failed to load Grinex orders (RUB): %w", err)
    }

    return parseGrinexHTML(html, pair, side)
}

func fetchGrinexOrdersUSDTA7A5(url string, pair domain.Pair, side domain.OrderSide, panelClass string, marketTab string) ([]*domain.Order, error) {
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
        return nil, fmt.Errorf("failed to load Grinex orders (A7A5): %w", err)
    }

    return parseGrinexHTML(html, pair, side)
}

func parseGrinexHTML(html string, pair domain.Pair, side domain.OrderSide) ([]*domain.Order, error) {
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
        return nil, fmt.Errorf("failed to parse HTML: %w", err)
    }

    var orders []*domain.Order
    rows := doc.Find("tbody tr")
    for i := 0; i < 5 && i < rows.Length(); i++ {
        row := rows.Eq(i)
        cols := row.Find("td")
        if cols.Length() < 3 {
            continue
        }

        price, _ := parseGrinexNumber(cols.Eq(0).Text())
        amount, _ := parseGrinexNumber(cols.Eq(1).Text())
        sum, _ := parseGrinexNumber(cols.Eq(2).Text())

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
        return nil, fmt.Errorf("no orders parsed from Grinex")
    }
    return orders, nil
}



func parseGrinexNumber(s string) (float64, error) {
	s = strings.ReplaceAll(s, "\u00a0", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	return strconv.ParseFloat(s, 64)
}