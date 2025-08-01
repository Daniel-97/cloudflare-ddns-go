package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

type Config struct {
	CLOUDFLARE_EMAIL   string
	CLOUDFLARE_API_KEY string
}

func main() {

	config, err := load_config()

	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	ipv4, err := get_current_ip()
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	cloudflare_overwrite_dns(*config, ipv4)
}

func cloudflare_overwrite_dns(config Config, ip string) {

	const url = "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records/$DNS_RECORD_ID"
	json := ""
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(json))

	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Email", config.CLOUDFLARE_EMAIL)
	req.Header.Set("X-Auth-Key", config.CLOUDFLARE_API_KEY)

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {

	}

	if res.StatusCode == http.StatusOK {

	} else {

	}

}

func load_config() (*Config, error) {
	var config = Config{
		CLOUDFLARE_EMAIL:   os.Getenv("CLOUDFLARE_EMAIL"),
		CLOUDFLARE_API_KEY: os.Getenv("CLOUDFLARE_API_KEY"),
	}

	if config.CLOUDFLARE_EMAIL == "" {
		return nil, errors.New("missing 'CLOUDFLARE_EMAIL' env")
	}

	if config.CLOUDFLARE_API_KEY == "" {
		return nil, errors.New("missing 'CLOUDFLARE_API_KEY' env")
	}

	return &config, nil
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
