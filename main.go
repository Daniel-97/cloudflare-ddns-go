package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	// Read config
	config := load_config()
	interval := time.Duration(config.REFRESH_INTERVAL) * time.Minute
	for {
		err := cloudflare_job(config)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(interval)
	}
}

func cloudflare_job(config *Config) error {
	ip_address, err := get_current_ip()
	if err != nil {
		return err
	}

	log.Println("Your ip address is", ip_address)

	record, err := cloudflare_get_dns_record(*config)

	if err != nil {
		return err
	} else if record != nil {
		// Record already present, overwrite it
		log.Printf("DNS record %s found in zone %s", config.CLOUDFLARE_RECORD_NAME, config.CLOUDFLARE_ZONE_ID)
		updated, err := cloudflare_update_dns_record(*config, record.ID, ip_address)

		if err != nil {
			return err
		} else if updated {
			log.Println("DNS record succesfully updated!")
		} else {
			log.Println("DNS record not updated")
		}
	} else {
		// Record do not exists, create a new one
		record_id, err := cloudflare_create_dns_record(*config, ip_address)
		if err != nil {
			return err
		}

		log.Printf("DNS record succesfully created with id: %s", record_id)
	}

	return nil
}

func get_dns_record_type(address string) string {
	if strings.Count(address, ":") >= 2 {
		return "AAAA"
	} else {
		return "A"
	}
}

func parse_cloudflare_response[T any](response *http.Response) (json_body T, err error) {

	var result T
	body, err := io.ReadAll(response.Body)

	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	err = json.Unmarshal(body, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func cloudflare_get_dns_record(config Config) (*DDNRecordResult, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", config.CLOUDFLARE_ZONE_ID)
	log.Println("Searching dns record", config.CLOUDFLARE_RECORD_NAME)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.CLOUDFLARE_API_TOKEN))

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	cloudflare_response, err := parse_cloudflare_response[CloudflareListRecordAPIResponse](res)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", res.StatusCode, cloudflare_response.Errors[0].Message)
	}

	log.Printf("Found %d dns records", len(cloudflare_response.Result))
	for _, record := range cloudflare_response.Result {
		if record.Name == config.CLOUDFLARE_RECORD_NAME {
			return &record, nil
		}
	}

	return nil, nil

}

func cloudflare_update_dns_record(config Config, record_id string, address string) (bool, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", config.CLOUDFLARE_ZONE_ID, record_id)

	req_body := DDNRecordResult{
		Name:    config.CLOUDFLARE_RECORD_NAME,
		TTL:     config.CLOUDFLARE_RECORD_TTL,
		Type:    get_dns_record_type(address),
		Comment: fmt.Sprintf("cloudflare-ddns-go (%s)", time.Now().Format(time.RFC3339)),
		Content: address,
		Proxied: config.CLOUDFLARE_RECORD_PROXY,
	}

	log.Printf("Updating Cloudflare dns %s record (%s) for address %s -> %s", req_body.Type, record_id, req_body.Content, req_body.Name)

	json_bytes, err := json.Marshal(req_body)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(json_bytes))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.CLOUDFLARE_API_TOKEN))

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return false, err
	}

	cloudflare_response, err := parse_cloudflare_response[CloudflareCreateRecordAPIResponse](res)
	if err != nil {
		return false, err
	}

	if res.StatusCode == http.StatusOK {
		return true, nil

	} else {
		return false, fmt.Errorf("HTTP error %d: %s", res.StatusCode, cloudflare_response.Errors[0].Message)
	}

}

func cloudflare_create_dns_record(config Config, address string) (record_id string, err error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", config.CLOUDFLARE_ZONE_ID)

	req_body := DDNRecordResult{
		Name:    config.CLOUDFLARE_RECORD_NAME,
		TTL:     config.CLOUDFLARE_RECORD_TTL,
		Type:    get_dns_record_type(address),
		Comment: fmt.Sprintf("cloudflare-ddns-go (%s)", time.Now().Format(time.RFC3339)),
		Content: address,
		Proxied: config.CLOUDFLARE_RECORD_PROXY,
	}
	log.Printf("Creating new Cloudflare dns %s record for address %s -> %s", req_body.Type, req_body.Content, req_body.Name)

	json_bytes, err := json.Marshal(req_body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_bytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.CLOUDFLARE_API_TOKEN))

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	cloudflare_response, err := parse_cloudflare_response[CloudflareCreateRecordAPIResponse](res)
	if err != nil {
		return "", err
	}

	if res.StatusCode == http.StatusOK {
		return cloudflare_response.Result.ID, nil

	} else {
		return "", fmt.Errorf("HTTP error %d: %s", res.StatusCode, cloudflare_response.Errors[0].Message)
	}

}

func load_config() *Config {
	var config = Config{
		CLOUDFLARE_API_TOKEN:    os.Getenv("CLOUDFLARE_API_TOKEN"),
		CLOUDFLARE_ZONE_ID:      os.Getenv("CLOUDFLARE_ZONE_ID"),
		CLOUDFLARE_RECORD_NAME:  os.Getenv("CLOUDFLARE_RECORD_NAME"),
		CLOUDFLARE_RECORD_TTL:   3600,
		CLOUDFLARE_RECORD_PROXY: false,
		REFRESH_INTERVAL:        5,
	}

	if config.CLOUDFLARE_API_TOKEN == "" {
		log.Fatal("missing 'CLOUDFLARE_API_TOKEN' env")
	}

	if config.CLOUDFLARE_ZONE_ID == "" {
		log.Fatal("missing 'CLOUDFLARE_ZONE_ID' env")
	}

	if config.CLOUDFLARE_RECORD_NAME == "" {
		log.Fatal("missing 'CLOUDFLARE_RECORD_NAME' env")
	}

	if os.Getenv("CLOUDFLARE_RECORD_TTL") != "" {
		ttl, err := strconv.Atoi(os.Getenv("CLOUDFLARE_RECORD_TTL"))
		if err != nil {
			log.Fatal("Invalid CLOUDFLARE_RECORD_TTL value")
		}
		config.CLOUDFLARE_RECORD_TTL = ttl
	}

	if os.Getenv("CLOUDFLARE_RECORD_PROXY") != "" {
		config.CLOUDFLARE_RECORD_PROXY, _ = strconv.ParseBool(os.Getenv("CLOUDFLARE_RECORD_PROXY"))
	}

	if os.Getenv("REFRESH_INTERVAL") != "" {
		interval, err := strconv.Atoi(os.Getenv("REFRESH_INTERVAL"))
		if err != nil {
			log.Fatal("Invalid REFRESH_INTERVAL value")
		}
		config.REFRESH_INTERVAL = interval
	}

	return &config
}

func get_current_ip() (ip string, err error) {

	log.Println("Looking for ip address...")
	url := "https://ifconfig.me/ip"
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "plain/text")

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {

	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	bodyString := string(body)

	return bodyString, nil

}
