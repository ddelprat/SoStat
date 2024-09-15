package objects

type Player struct {
	FirstName     string    `json:"first_name,firstName"`
	LastName      string    `json:"last_name,lastName"`
	Age           int       `json:"age"`
	Slug          string    `json:"slug"`
	Team          string    `json:"team"`
	League        string    `json:"league"`
	Status        string    `json:"status"`
	So5Scores     []float32 `json:"so5_scores,so5Scores"`
	AverageScore  float32   `json:"average_score,averageScore"`
	LastPrices    []int     `json:"last_prices,lastPrices"`
	PlayingStatus float32   `json:"playing_status,playingStatus"`
}
