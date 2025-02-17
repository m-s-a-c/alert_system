package box

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/m-s-a-c/alert_system.git/core/config"
)

func ZboxReplication(provL []map[string]interface{}) (string, bool) {
	var boxUrl = "https://0box." + config.Configuration.NetworkSubdomain + "." + config.Configuration.NetworkDomain + "/v2/latest-snapshot"

	req, err := http.NewRequest("GET", boxUrl, nil)
	if err != nil {
		log.Fatalf("Failed to fetch 0box url: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perfrom GET request of type application/json on 0box url:  %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: Recieved status code for 0box url: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read the response body from 0box url: %v", err)
	}

	if len(body) == 0 {
		log.Fatal("Failed to read the response body for 0box url")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error while parsing JSON:", err)
	}

	var conc []int

	for _, rMap := range provL {
		intRound, _ := strconv.Atoi(rMap["round"].(string))
		conc = append(conc, intRound)
	}

	maxRound := slices.Max(conc)

	roundDiff := maxRound - int(data["round"].(float64))

	if roundDiff > 20 {
		log.Printf("ALERT: 0BOX REPLICATION IS BEHIND BY %d ROUNDS!", roundDiff)
		return fmt.Sprintf("%s replication is behind by %d rounds", boxUrl, roundDiff), false
	}

	log.Println("INFO: 0BOX REPLICATION IS WORKING FINE")
	return fmt.Sprintf("%s replication is in sync", boxUrl), true
}
