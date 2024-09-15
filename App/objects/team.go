package objects

type Team struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	DomesticLeague string `json:"domesticLeague"`
	Country        string `json:"country"`
}
