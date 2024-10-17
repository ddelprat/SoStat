package rating

import (
	"context"
	"main/App/objects"
	"sort"
)

var Rating_reason_translations = map[string]string{
	"recent_price": "Price lower than average last sales",
	"live_price":   "Price lower than current live prices",
}

type Rating struct {
	ctx context.Context
}

func NewRating() *Rating {
	return &Rating{
		ctx: context.Background(),
	}
}

// RateCard gives a score to the card between 0 and 100 the more the score is high, the better is the trade.
func (cr *Rating) RateCard(card *objects.Sale, liveSales []*objects.Sale, lastSales []*objects.Sale) map[string]float64 {
	scores := make(map[string]float64)
	// if there is not enough data it is not relevant.
	if len(lastSales) <= 5 {
		return nil
	}

	//compute different scores
	scores["recent_price"] = cr.compareToRecentPrice(card, lastSales)
	scores["current_price"] = cr.compareToLivePrices(card, liveSales)
	//scores["price_trend"] = cr.computePriceTrend(card, lastSales)

	return scores
}

// compareToRecentPrice rates the card depending on its price difference with the average of last prices.
// 1 - price/average + (1 - price/min)/10
func (cr *Rating) compareToRecentPrice(card *objects.Sale, lastSales []*objects.Sale) float64 {
	var averagePrice uint64
	minPrice := lastSales[0].Price
	for _, c := range lastSales {
		averagePrice += c.Price
		if c.Price < minPrice {
			minPrice = c.Price
		}
	}
	minPriceRatio := float64(card.Price) / float64(minPrice)
	if minPriceRatio > 1.2 {
		return 0
	}

	averagePrice = averagePrice / uint64(len(lastSales))
	return max(float64(1)-float64(card.Price)/float64(averagePrice)+(float64(1)-minPriceRatio)/10, 0)
}

// compareToLivePrices rates the card depending on its price difference with the average of last prices.
func (cr *Rating) compareToLivePrices(card *objects.Sale, liveSales []*objects.Sale) float64 {
	if len(liveSales) < 2 {
		return 0
	}
	// sort by ascending price
	sort.Slice(liveSales, func(i, j int) bool {
		return liveSales[i].Price < liveSales[j].Price
	})
	if liveSales[1].Price == 0 {
		return 0
	}
	return max(float64(1)-float64(card.Price)/float64(liveSales[1].Price), 0)
}

// computePriceTrend rates the card depending on the price trend (descending or ascending).
func (cr *Rating) computePriceTrend(lastSales []*objects.Sale) float64 {
	// sort by date
	sort.Slice(lastSales, func(i, j int) bool {
		return lastSales[i].Date.Before(lastSales[j].Date)
	})

	var meanDerivative float64
	for i := 0; i < len(lastSales)-1; i++ {
		meanDerivative += float64(lastSales[i+1].Price - lastSales[i].Price) // difference between consecutive elements
	}
	meanDerivative = meanDerivative / float64(len(lastSales)*1e18)
	return 0
}
