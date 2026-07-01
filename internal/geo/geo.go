package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Location struct {
	Country string
	City    string
}

var client = &http.Client{Timeout: 3 * time.Second}

func Lookup(ip string) Location {
	resp, err := client.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=country,city", ip))
	if err != nil {
		return Location{}
	}
	defer resp.Body.Close()

	var result struct {
		Country string `json:"country"`
		City    string `json:"city"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return Location{Country: result.Country, City: result.City}
}
