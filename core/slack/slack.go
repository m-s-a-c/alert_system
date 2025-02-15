package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SlackMessage represents the payload structure
type SlackMessage struct {
	Text        string       `json:"text"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment for additional formatting
type Attachment struct {
	Color string `json:"color"`
	Text  string `json:"text"`
}

func SendSlackMessage(msg []interface{}, alertype, provType, webhookURL, channel string) error {
	var msgBuilder strings.Builder
	var textMessage string

	for _, v := range msg {
		msgBuilder.WriteString(fmt.Sprintf("❌ %v \n", v))
	}

	if len(msg) == 0 {
		return nil
	}else {
		if alertype == "unreachable" {
			textMessage = "📝 BELOW " + provType + " ARE NOT REACHABLE. 📝\n"
		}else {
			textMessage = "📝 BELOW " + provType + " ARE LAGGING. 📝\n"
		}
	}

	payload := SlackMessage{
		Text:     textMessage,
		Username: "Mainnet Alerts", // Customize bot name
		//IconEmoji: ":robot:",        // Customize bot icon
		Channel: channel, // Channel (optional if Webhook is tied to a specific one)
		Attachments: []Attachment{
			{Color: "#f50000", Text: msgBuilder.String()},
		},
	}

	// Convert struct to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Make HTTP request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API responded with status: %d", resp.StatusCode)
	}
	return nil
}
