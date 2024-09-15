package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
)

type Notifications struct {
	ctx context.Context
}

func NewNotifications() *Notifications {
	return &Notifications{
		ctx: context.Background(),
	}
}

func (n *Notifications) getUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", os.Getenv("TELEGRAM_BOT_TOKEN"))
}

func (n *Notifications) SendTelegramMessage(text string) (bool, error) {
	// Global variables
	var err error
	var response *http.Response

	err = godotenv.Load()
	if err != nil {
		return false, nil
	}

	// Send the message
	url := fmt.Sprintf("%s/sendMessage", n.getUrl())
	body, _ := json.Marshal(map[string]string{
		"chat_id": os.Getenv("TELEGRAM_CHAT_ID"),
		"text":    text,
	})
	response, err = http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return false, err
	}

	// Close the request at the end
	defer response.Body.Close()

	// Body
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	return true, nil
}
