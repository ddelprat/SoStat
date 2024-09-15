package scout

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
	"main/App/notifications"
	"slices"
	"strconv"
)

var premierLeagueTeams = []string{"afc-bournemouth-bournemouth-dorset", "arsenal-london", "aston-villa-birmingham", "brentford-brentford-middlesex", "brighton-hove-albion-brighton-east-sussex",
	"chelsea-london", "crystal-palace-london", "everton-liverpool", "fulham-london", "ipswich-town-ipswich-suffolk", "liverpool-liverpool", "leicester-city-leicester", "manchester-city-manchester",
	"manchester-united-manchester", "newcastle-united-newcastle-upon-tyne", "nottingham-forest-nottingham", "southampton-southampton-hampshire", "tottenham-hotspur-london",
	"west-ham-united-london", "wolverhampton-wanderers-wolverhampton"}

type Card struct {
	Slug       string
	playerSlug string
	FirstName  string
	LastName   string
	Season     int64
	Price      uint64
}

type Scout struct {
	ctx               context.Context
	client            *graphql.Client
	apiKey            string
	GraphqlQueryCount int64
	notif             *notifications.Notifications
}

func NewScout(client *graphql.Client, apiKey string) *Scout {
	return &Scout{
		ctx:               context.Background(),
		client:            client,
		apiKey:            apiKey,
		GraphqlQueryCount: 0,
		notif:             notifications.NewNotifications(),
	}
}

func (s *Scout) ScoutMarket() error {
	cardsToBuy := make([]Card, 0)
	scoutedPlayers := make(map[string]bool)
	for _, team := range premierLeagueTeams {
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

			lastSales, err := s.buildLiveSalesList(player)
			if err != nil {
				return err
			}

			cardsToBuy = append(cardsToBuy, s.computeCardsToBuy(liveSales, lastSales)...)
		}
	}

	for _, card := range cardsToBuy {
		message := fmt.Sprintf(
			"Player: %s %s \n"+
				"Season : %d \n"+
				"price: %.4f ETH \n"+
				"https://sorare.com/fr/football/players/%s/cards?card=%s", card.FirstName, card.LastName, card.Season, float64(card.Price)/float64(1000000000000000000), card.playerSlug, card.Slug)
		if _, err := s.notif.SendTelegramMessage(message); err != nil {
			return err
		}

	}

	return nil
}

func (s *Scout) buildLiveSalesList(player string) (map[int64][]Card, error) {
	graphqlLiveSales, err := s.listPlayerLiveSales(player)
	if err != nil {
		return nil, err
	}
	s.GraphqlQueryCount += 1

	liveSales := make(map[int64][]Card)
	for _, sale := range graphqlLiveSales.Tokens.LiveSingleSaleOffers.Nodes {
		if len(sale.SenderSide.AnyCards) != 1 {
			continue
		}
		if sale.SenderSide.AnyCards[0].RarityTyped == "limited" {
			price, _ := strconv.ParseUint(sale.ReceiverSide.Wei, 10, 64)
			liveSales[sale.SenderSide.AnyCards[0].SeasonYear] = append(liveSales[sale.SenderSide.AnyCards[0].SeasonYear], Card{
				Slug:       sale.SenderSide.AnyCards[0].Slug,
				playerSlug: player,
				FirstName:  sale.SenderSide.AnyCards[0].AnyPlayer.FirstName,
				LastName:   sale.SenderSide.AnyCards[0].AnyPlayer.LastName,
				Season:     sale.SenderSide.AnyCards[0].SeasonYear,
				Price:      price,
			})
		}
	}

	return liveSales, nil
}

func (s *Scout) buildLastSalesList(player string) (map[int64][]Card, error) {
	graphqlLastSales, err := s.listLastPlayerSales(player, "")
	if err != nil {
		return nil, err
	}
	s.GraphqlQueryCount += 1
	lastSales := make(map[int64][]Card)
	for _, sale := range graphqlLastSales.Football.Player.TokenPrices.Nodes {
		price, _ := strconv.ParseUint(sale.Amounts.Wei, 10, 64)
		lastSales[sale.Card.SeasonYear] = append(lastSales[sale.Card.SeasonYear], Card{
			Slug:  sale.Card.Slug,
			Price: price,
		})
	}

	return lastSales, nil
}

func (s *Scout) computeCardsToBuy(liveSales map[int64][]Card, lastSales map[int64][]Card) []Card {
	res := make([]Card, 0)
	for season, cards := range liveSales {
		if _, ok := lastSales[season]; !ok {
			continue
		}
		minPriceCard := slices.MinFunc(cards, func(a, b Card) int {
			return int(a.Price - b.Price)
		})

		var averagePrice uint64
		for _, card := range lastSales[season] {
			averagePrice += card.Price
		}
		averagePrice = averagePrice / uint64(len(lastSales[season]))
		if minPriceCard.Price < uint64(0.9*float64(averagePrice)) {
			res = append(res, minPriceCard)
		}
	}

	return res
}
