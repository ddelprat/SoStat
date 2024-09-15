package scout

import (
	"fmt"
	"github.com/machinebox/graphql"
	"log"
	"main/App/objects"
)

type GetPlayerResponse struct {
	Football struct {
		Player struct {
			ActiveClub struct {
				Name string `json:"name"`
			} `json:"activeClub"`
			Position      string  `json:"position"`
			FirstName     string  `json:"firstName"`
			LastName      string  `json:"lastName"`
			Age           int     `json:"age"`
			Slug          string  `json:"slug"`
			AverageScore  float64 `json:"averageScore"`
			PlayingStatus string  `json:"playingStatus"`
			So5Scores     []struct {
				Score float64 `json:"score"`
			} `json:"so5Scores"`
		} `json:"player"`
	} `json:"football"`
}

type goScoutResponse struct {
	Data struct {
		AllCards struct {
			Nodes []struct {
				Player objects.Player `json:"player"`
			} `json:"nodes"`
		} `json:"allCards"`
	} `json:"data"`
}

// getPlayer gets all the info about a player.
func (s *Scout) GetPlayer(player string) (*GetPlayerResponse, error) {
	if player == "" {
		return nil, fmt.Errorf("player slug cannot be empty")
	}
	request := graphql.NewRequest(fmt.Sprintf(`
		query{
			football{
				player(slug:"%s"){
					activeClub{
						name
					}
					position
					firstName
					lastName
					age
					slug
					averageScore(type: LAST_FIFTEEN_SO5_AVERAGE_SCORE)
					playingStatus
					so5Scores(last: 5) {
						score
					}
				}
			}
		}`, player))
	request.Header.Set("APIKEY", s.apiKey)
	var res *GetPlayerResponse

	// Run the request
	if err := s.client.Run(s.ctx, request, &res); err != nil {
		log.Fatal(err)
	}

	return res, nil
}

// goScout scout any player corresponding to a given position and league.
func (s *Scout) scoutPlayer(position string, league string) (*goScoutResponse, error) {
	if position == "" || league == "" {
		return nil, fmt.Errorf("position or league is empty")
	}
	rawQuery := fmt.Sprintf(`
		query {
			allCards(rarities:rare, positions:%s, teamSlugs:"%s", first:100) {
				nodes {
					player {
						firstName
						lastName
						age
						slug
						averageScore(type: LAST_FIFTEEN_SO5_AVERAGE_SCORE)
						playingStatus
						so5Scores(last: 5) {
						score
					}
				}
			}
		}
	}`, position, league)

	var res *goScoutResponse
	// Run the request
	if err := s.client.Run(s.ctx, graphql.NewRequest(rawQuery), &res); err != nil {
		log.Fatal(err)
	}

	return res, nil
}
