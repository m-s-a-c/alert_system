package chain

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"
)

func GetProvidersURL(url string) *Providers {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to fetch blockworker url: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform GET request of type application/json on blockworker url: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: Recieved status code for blockworker url: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read the response body for blockworker url: %v", err)
	}
	if len(body) == 0 {
		log.Fatalf("Failed to get providers from the provided url")
	}

	var prov Providers

	if err = json.Unmarshal(body, &prov); err != nil {
		log.Fatalf("Unable to unmarshall json for providers %v", err)
	}

	return &prov
}

func CheckUrlStatus(provL []interface{}) ([]interface{}, []interface{}) {
	if len(provL) == 0 {
		fmt.Println("No providers received in checkUrlStatus")
		return nil, nil
	}

	var (
		inValidProvL []interface{}
		validProvL   []interface{}
		mutex        sync.Mutex
		wg           sync.WaitGroup
		client       = &http.Client{Timeout: 10 * time.Second} // Set timeout
	)

	wg.Add(len(provL))

	for _, url := range provL {
		go func(url string) {
			defer wg.Done()
			resp, err := client.Get(url)
			if err != nil || resp.StatusCode != http.StatusOK {
				mutex.Lock()
				inValidProvL = append(inValidProvL, err)
				mutex.Unlock()
				return
			}
			mutex.Lock()
			validProvL = append(validProvL, url)
			mutex.Unlock()
			resp.Body.Close()
		}(url.(string))

	}
	wg.Wait()
	return validProvL, inValidProvL
}

func CheckProviderRound(provL []interface{}) ([]map[string]interface{}, error) {
	if len(provL) == 0 {
		fmt.Println("No sharder/miner urls passed")
		return nil, nil
	}

	var (
		client  = &http.Client{Timeout: 10 * time.Second} // Set timeout
		mutex   sync.Mutex
		wg      sync.WaitGroup
		results []map[string]interface{}
	)

	wg.Add(len(provL))

	for _, url := range provL {
		go func(url string) {
			defer wg.Done()
			resp, err := client.Get(url + "/_diagnostics/round_info")
			if err != nil || resp.StatusCode != http.StatusOK {
				fmt.Printf("Url %s is not reachable %v\n", url, err)
				return
			}
			defer resp.Body.Close()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Failed to read response from %s: %v\n", url, err)
				return
			}

			// Regular expression to extract "Round: <number>"
			re := regexp.MustCompile(`Round:\s*(\d+)`)
			match := re.FindStringSubmatch(string(body))

			if len(match) > 1 {
				// Store the result in the interface
				mutex.Lock()
				results = append(results, map[string]interface{}{
					"url":   url,
					"round": match[1],
				})
				mutex.Unlock()
			}
		}(url.(string))

	}

	wg.Wait()

	if len(results) == 0 {
		return nil, fmt.Errorf("round number not found in any response")
	}

	return results, nil
}

func LaggingProviders(provL []map[string]interface{}) []interface{} {
	conc := []int{}

	for _, rMap := range provL {
		intRound, _ := strconv.Atoi(rMap["round"].(string))
		conc = append(conc, intRound)
	}

	maxRound := slices.Max(conc)
	var data []interface{}
	for _, rUrl := range provL {
		ro := rUrl["round"].(string)
		amm3, _ := strconv.Atoi(ro)
		if amm3 < (maxRound - 100) {
			data = append(data, rUrl["url"].(string)+" is behind the current round at "+rUrl["round"].(string))
		}
	}

	return data
}
