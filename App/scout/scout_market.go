package scout

import (
	"fmt"
	"github.com/machinebox/graphql"
	"time"
)

type listLastPlayerSalesResponse struct {
	Football struct {
		Player struct {
			TokenPrices struct {
				Nodes []struct {
					ID   string    `json:"id"`
					Date time.Time `json:"date"`
					Card struct {
						Slug       string `json:"slug"`
						SeasonYear int64  `json:"seasonYear"`
					} `json:"card"`
					Amounts struct {
						EUR int    `json:"eur"`
						Wei string `json:"wei"`
					} `json:"amounts"`
				} `json:"nodes"`
				PageInfo struct {
					StartCursor string `json:"startCursor"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"tokenPrices"`
		} `json:"player"`
	} `json:"football"`
}

type listPlayerLiveSalesResponse struct {
	Tokens struct {
		LiveSingleSaleOffers struct {
			TotalCount int `json:"totalCount"`
			Nodes      []struct {
				CreatedAt    string `json:"createdAt"`
				ReceiverSide struct {
					Wei string `json:"wei"`
				} `json:"receiverSide"`
				SenderSide struct {
					AnyCards []struct {
						AnyPlayer struct {
							FirstName string `json:"firstName"`
							LastName  string `json:"lastName"`
						} `json:"anyPlayer"`
						Slug        string `json:"slug"`
						RarityTyped string `json:"rarityTyped"`
						SeasonYear  int64  `json:"SeasonYear"`
					} `json:"anyCards"`
				} `json:"senderSide"`
			} `json:"nodes"`
			PageInfo struct {
				EndCursor string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"liveSingleSaleOffers"`
	} `json:"tokens"`
}

// listLastPlayerSales lists the 50 last sales for a given player.
func (s *Scout) listLastPlayerSales(player string, cursor string) (*listLastPlayerSalesResponse, error) {
	if player == "" {
		return nil, fmt.Errorf("team slug cannot be empty")
	}

	request := graphql.NewRequest(`
	playerLastSales($player: String!, $cursor: String!) {
		query {
		  football{
			player(slug:$player){
			  tokenPrices(rarity:limited, after:$cursor){
				nodes{
				  id
				  date
				  card{
					slug
					seasonYear
				  }
				  amounts{
					eur
					wei
				  }
				}
				pageInfo{
				  startCursor
				  endCursor
				}
			  }
			}
		  }
		}
	}`)

	request.Var("player", player)
	request.Var("cursor", cursor)
	request.Header.Set("APIKEY", s.apiKey)

	var res listLastPlayerSalesResponse

	// Run the request
	if err := s.client.Run(s.ctx, request, &res); err != nil {
		return nil, fmt.Errorf("failed to run market query: %w", err)
	}

	return &res, nil
}

// listPlayerLiveSales lists the live sales offers for a given player.
func (s *Scout) listPlayerLiveSales(player string) (*listPlayerLiveSalesResponse, error) {
	if player == "" {
		return nil, fmt.Errorf("team slug cannot be empty")
	}

	request := graphql.NewRequest(`
		query getPlayerLiveSales($player: String!) {
		  tokens {
			liveSingleSaleOffers(sport: FOOTBALL, playerSlug: $player) {
			  totalCount
			  nodes {
				createdAt
				receiverSide {
				  wei
				}
				senderSide {
				  anyCards {
					anyPlayer{
					  firstName
					  lastName
					}
					slug
					rarityTyped
					seasonYear
				  }
				}
			  }
			  pageInfo {
				endCursor
			  }
			}
		  }
		}`)

	request.Var("player", player)
	request.Header.Set("APIKEY", s.apiKey)

	var res listPlayerLiveSalesResponse

	// Run the request
	if err := s.client.Run(s.ctx, request, &res); err != nil {
		return nil, fmt.Errorf("failed to run market query: %w", err)
	}

	return &res, nil
}
