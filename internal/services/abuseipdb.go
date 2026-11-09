package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// AbuseResponse memetakan struktur JSON dari AbuseIPDB
type AbuseResponse struct {
	Data struct {
		IPAddress            string `json:"ipAddress"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
		CountryCode          string `json:"countryCode"`
		UsageType            string `json:"usageType"`
		ISP                  string `json:"isp"`
		Domain               string `json:"domain"`
		TotalReports         int    `json:"totalReports"`
	} `json:"data"`
}

// CheckIPReputation memanggil API AbuseIPDB untuk mendapatkan skor ancaman (0-100)
func CheckIPReputation(ip string) (*AbuseResponse, error) {
	apiKey := os.Getenv("ABUSEIPDB_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ABUSEIPDB_KEY belum diatur di .env")
	}

	url := "https://api.abuseipdb.com/api/v2/check?ipAddress=" + ip + "&maxAgeInDays=90"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Key", apiKey)
	req.Header.Add("Accept", "application/json")

	// Set timeout 10 detik agar aplikasi tidak hang jika API sedang lambat
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gagal mengecek IP, status HTTP: %d", resp.StatusCode)
	}

	var result AbuseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}