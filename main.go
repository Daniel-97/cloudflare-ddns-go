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

type Config struct {
	CLOUDFLARE_API_TOKEN   string
	CLOUDFLARE_ZONE_ID     string
	CLOUDFLARE_RECORD_NAME string
}

type CloudflareAPIRequest struct {
	Name    string `json:"name"`
	TTL     int    `json:"ttl"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

type CloudflareAPIResponse struct {
	Errors   []APIMessage `json:"errors"`
	Messages []APIMessage `json:"messages"`
	Success  bool         `json:"success"`
	Result   Result       `json:"result"`
}

type APIMessage struct {
	Code             int    `json:"code"`
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	Source           Source `json:"source"`
}

type Source struct {
	Pointer string `json:"pointer"`
}

type Result struct {
	Name     string   `json:"name"`
	TTL      int      `json:"ttl"`
	Type     string   `json:"type"`
	Comment  string   `json:"comment"`
	Content  string   `json:"content"`
	Proxied  bool     `json:"proxied"`
	Settings Settings `json:"settings"`
	Tags     []string `json:"tags"`
	ID       string   `json:"id"`
}

type Settings struct {
	IPv4Only bool `json:"ipv4_only"`
	IPv6Only bool `json:"ipv6_only"`
}

func main() {

	config := load_config()

	ip, err := get_current_ip()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Your ip address is", ip)

	record_id, err := cloudflare_create_dns_record(*config, ip)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("DNS record succesfully created with id: %s", record_id)
}

func get_dns_record_type(address string) string {
	if strings.Count(address, ":") >= 2 {
		return "AAAA"
	} else {
		return "A"
	}
}

func parse_cloudflare_response(response *http.Response) (json_body CloudflareAPIResponse, err error) {

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return CloudflareAPIResponse{}, err
	}

	var apiResp CloudflareAPIResponse
	err = json.Unmarshal(body, &apiResp)

	if err != nil {
		return CloudflareAPIResponse{}, err
	}

	return apiResp, nil
}

func cloudflare_get_dns_record(config Config, record_name: str) {
	
}
func cloudflare_create_dns_record(config Config, address string) (record_id string, err error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", config.CLOUDFLARE_ZONE_ID)
	log.Println(url)

	req_body := CloudflareAPIRequest{
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

	cloudflare_response, err := parse_cloudflare_response(res)
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
