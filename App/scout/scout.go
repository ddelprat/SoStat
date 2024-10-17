package scout

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
	"main/App/notifications"
	"main/App/objects"
	"main/App/rating"
	"sort"
	"strconv"
	"time"
)

var premierLeagueTeams = []string{"afc-bournemouth-bournemouth-dorset", "arsenal-london", "aston-villa-birmingham", "brentford-brentford-middlesex", "brighton-hove-albion-brighton-east-sussex",
	"chelsea-london", "crystal-palace-london", "everton-liverpool", "fulham-london", "ipswich-town-ipswich-suffolk", "liverpool-liverpool", "leicester-city-leicester", "manchester-city-manchester",
	"manchester-united-manchester", "newcastle-united-newcastle-upon-tyne", "nottingham-forest-nottingham", "southampton-southampton-hampshire", "tottenham-hotspur-london",
	"west-ham-united-london", "wolverhampton-wanderers-wolverhampton"}

var testTeams = []string{"arsenal-london"}

type cardRating struct {
	card   *objects.Sale
	scores map[string]float64
}

type Scout struct {
	ctx               context.Context
	client            *graphql.Client
	apiKey            string
	GraphqlQueryCount int64
	notif             *notifications.Notifications
	cr                *rating.Rating
}

func NewScout(client *graphql.Client, apiKey string) *Scout {
	return &Scout{
		ctx:               context.Background(),
		client:            client,
		apiKey:            apiKey,
		GraphqlQueryCount: 0,
		notif:             notifications.NewNotifications(),
		cr:                rating.NewRating(),
	}
}

func (s *Scout) ScoutMarket() error {
	cardsToBuy := make(map[string]cardRating)
	scoutedPlayers := make(map[string]bool)
	for _, team := range testTeams {
		graphqlTeam, err := s.getTeam(team)
		if err != nil {
			return err
		}
		s.GraphqlQueryCount += 1

		playersToScout := make([]string, 0, len(graphqlTeam.Football.Club.Players.Nodes))
		for _, p := range graphqlTeam.Football.Club.Players.Nodes {
			if p.PlayingStatus == "RETIRED" || p.PlayingStatus == "NOT_PLAYING" {
				continue
			}
			playersToScout = append(playersToScout, p.Slug)
		}

		for _, player := range playersToScout {
			if _, ok := scoutedPlayers[player]; ok {
				continue
			}
			scoutedPlayers[player] = true
			liveSales, err := s.buildLiveSalesList(player)
			if err != nil {
				return err
			}
			if len(liveSales) == 0 {
				continue
			}

			lastSales, err := s.buildLastSalesList(player)
			if err != nil {
				return err
			}

			cardsRating := s.getCardsRating(liveSales, lastSales)
			for cardID := range cardsRating {
				cardsToBuy[cardID] = cardRating{
					card:   cardsRating[cardID].card,
					scores: cardsRating[cardID].scores,
				}
			}
		}
	}

	for _, c := range cardsToBuy {
		maxReason, maxScore := s.getMaxScore(c.scores)
		if maxScore < 0.65 {
			continue
		}
		message := fmt.Sprintf(
			"Player: %s %s \n"+
				"Season : %d \n"+
				"price: %.4f ETH \n"+
				"triggered: %s \n"+
				"score: %d \n"+
				"https://sorare.com/fr/football/players/%s/cards?card=%s", c.card.FirstName, c.card.LastName, c.card.Season, float64(c.card.Price)/float64(1000000000000000000), maxReason, int64(maxScore*100), c.card.PlayerSlug, c.card.Slug)
		if _, err := s.notif.SendTelegramMessage(message); err != nil {
			return err
		}

	}

	return nil
}

func (s *Scout) buildLiveSalesList(player string) (map[int64][]*objects.Sale, error) {
	graphqlLiveSales, err := s.listPlayerLiveSales(player)
	if err != nil {
		return nil, err
	}
	s.GraphqlQueryCount += 1

	liveSales := make(map[int64][]*objects.Sale)
	for _, sale := range graphqlLiveSales.Tokens.LiveSingleSaleOffers.Nodes {
		if len(sale.SenderSide.AnyCards) != 1 {
			continue
		}
		if sale.SenderSide.AnyCards[0].RarityTyped == "limited" {
			price, _ := strconv.ParseUint(sale.ReceiverSide.Wei, 10, 64)
			liveSales[sale.SenderSide.AnyCards[0].SeasonYear] = append(liveSales[sale.SenderSide.AnyCards[0].SeasonYear], &objects.Sale{
				Slug:       sale.SenderSide.AnyCards[0].Slug,
				PlayerSlug: player,
				FirstName:  sale.SenderSide.AnyCards[0].AnyPlayer.FirstName,
				LastName:   sale.SenderSide.AnyCards[0].AnyPlayer.LastName,
				Season:     sale.SenderSide.AnyCards[0].SeasonYear,
				Price:      price,
			})
		}
	}

	return liveSales, nil
}

func (s *Scout) buildLastSalesList(player string) (map[int64][]*objects.Sale, error) {
	currentTime := time.Now().UTC()
	lastMonth := currentTime.AddDate(0, -1, 0).Format(time.RFC3339)

	graphqlLastSales, err := s.listLastPlayerSales(player, lastMonth, "")
	if err != nil {
		return nil, err
	}
	s.GraphqlQueryCount += 1
	lastSales := make(map[int64][]*objects.Sale)
	for _, sale := range graphqlLastSales.Football.Player.TokenPrices.Nodes {
		price, _ := strconv.ParseUint(sale.Amounts.Wei, 10, 64)
		lastSales[sale.Card.SeasonYear] = append(lastSales[sale.Card.SeasonYear], &objects.Sale{
			Slug:  sale.Card.Slug,
			Price: price,
			Date:  sale.Date,
		})
	}

	return lastSales, nil
}

func (s *Scout) getCardsRating(liveSales map[int64][]*objects.Sale, lastSales map[int64][]*objects.Sale) map[string]cardRating {
	res := make(map[string]cardRating)
	season := int64(2024)
	classicLiveSales, seasonLiveSales := s.splitSeasonCards(liveSales, season)
	classicLastSales, seasonLastSales := s.splitSeasonCards(lastSales, season)

	// classic season
	minClassicPriceCard := s.checkMinLiveMarketPrice(classicLiveSales, 0.9)
	if minClassicPriceCard != nil {
		res[minClassicPriceCard.Slug] = cardRating{
			card:   minClassicPriceCard,
			scores: s.cr.RateCard(minClassicPriceCard, classicLiveSales, classicLastSales),
		}
	}

	// in season
	minSeasonPriceCard := s.checkMinLiveMarketPrice(seasonLiveSales, 0.9)
	if minSeasonPriceCard != nil {
		if s.cr == nil {
			s.cr = rating.NewRating()
		}
		res[minSeasonPriceCard.Slug] = cardRating{
			card:   minSeasonPriceCard,
			scores: s.cr.RateCard(minSeasonPriceCard, classicLiveSales, seasonLastSales),
		}
	}

	return res
}

// checkMinLiveMarketPrice checks if there is a card cheaper than the other on the live market by a certain rate and returns it, ie min_card_price < rate * second_min_card_price
func (s *Scout) checkMinLiveMarketPrice(cards []*objects.Sale, rate float64) *objects.Sale {
	if len(cards) < 2 {
		return nil
	}
	// sort by ascending price
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Price < cards[j].Price
	})
	if cards[0].Price < uint64(rate*float64(cards[1].Price)) {
		return cards[0]
	}
	return nil
}

// splitSeasonCards split cards between in-season and classic season cards.
func (s *Scout) splitSeasonCards(sales map[int64][]*objects.Sale, seasonYear int64) ([]*objects.Sale, []*objects.Sale) {
	classicCards := make([]*objects.Sale, 0)
	for season, cards := range sales {
		if season == seasonYear {
			continue
		}
		classicCards = append(classicCards, cards...)
	}
	return classicCards, sales[seasonYear]
}

// getMaxScore get the max score of the card and the reason for it.
func (s *Scout) getMaxScore(scores map[string]float64) (string, float64) {
	var maxReason string
	var maxScore float64
	for reason, score := range scores {
		if score > maxScore {
			maxScore = score
			maxReason = reason
		}
	}
	maxReason = rating.Rating_reason_translations[maxReason]
	return maxReason, maxScore
}
