package scout

import (
	"fmt"
	"log"

	"github.com/machinebox/graphql"
)

type getTeamResponse struct {
	Football struct {
		Club struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Players struct {
				Nodes []struct {
					PlayingStatus string `json:"playingStatus"`
					Slug          string `json:"slug"`
				} `json:"nodes"`
			} `json:"players"`
			Country struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"country"`
			DomesticLeague struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"domesticLeague"`
		} `json:"club"`
	} `json:"football"`
}

// getTeam get all the info about a team.
func (s *Scout) getTeam(team string) (*getTeamResponse, error) {
	if team == "" {
		return nil, fmt.Errorf("player slug cannot be empty")
	}
	rawQuery := fmt.Sprintf(`
		query{
			football{
				club(slug:"%s"){
					id
					name
					players{
					  nodes{
						playingStatus
						slug
					  }
					}
					country{
					id
					name
					slug
					}
					domesticLeague{
						id
					name
					}
				}
			}
		}`, team)

	var res *getTeamResponse

	// Run the request
	if err := s.client.Run(s.ctx, graphql.NewRequest(rawQuery), &res); err != nil {
		log.Fatal(err)
	}

	return res, nil
}
