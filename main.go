package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {

	config := load_config()

	ip, err := get_current_ip()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Your ip address is", ip)

	record, err := cloudflare_get_dns_record(*config, config.CLOUDFLARE_RECORD_NAME)

	if err != nil {
		log.Fatal(err)
	} else if record != nil {
		// Record already present, overwrite it
		log.Printf("Record %s found in zone %s", config.CLOUDFLARE_RECORD_NAME, config.CLOUDFLARE_ZONE_ID)
	} else {
		// Record do not exists, create a new one
		record_id, err := cloudflare_create_dns_record(*config, ip)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("DNS record succesfully created with id: %s", record_id)
	}
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

func cloudflare_get_dns_record(config Config, record_name string) (*DDNRecordResult, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", config.CLOUDFLARE_ZONE_ID)
	log.Println("Searching dns record", record_name)

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
		if record.Name == record_name {
			return &record, nil
		}
	}

	return nil, nil

}

func cloudflare_create_dns_record(config Config, address string) (record_id string, err error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", config.CLOUDFLARE_ZONE_ID)

	req_body := DDNRecordResult{
		Name:    config.CLOUDFLARE_RECORD_NAME,
		TTL:     3600,
		Type:    get_dns_record_type(address),
		Comment: "cloudflare-ddns-go",
		Content: address,
		Proxied: false,
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
		CLOUDFLARE_API_TOKEN:   os.Getenv("CLOUDFLARE_API_TOKEN"),
		CLOUDFLARE_ZONE_ID:     os.Getenv("CLOUDFLARE_ZONE_ID"),
		CLOUDFLARE_RECORD_NAME: os.Getenv("CLOUDFLARE_RECORD_NAME"),
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
