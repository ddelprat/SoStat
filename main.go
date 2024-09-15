package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/machinebox/graphql"
	"log"
	"main/App/scout"
	"net/http"
	"os"
)

var apiKey string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey = os.Getenv("SORARE_API_KEY")
	// Define routes
	http.HandleFunc("/scout", scoutHandler)

	// Start the web server
	fmt.Println("Server is running on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func scoutHandler(_ http.ResponseWriter, _ *http.Request) {
	client := graphql.NewClient("https://api.sorare.com/graphql/")
	scouting := scout.NewScout(client, apiKey)
	err := scouting.ScoutMarket()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Total number of requests : ", scouting.GraphqlQueryCount)
}
