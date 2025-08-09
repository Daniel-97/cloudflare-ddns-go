package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloudflare-ddns-go/cloudflare"
)

type Config struct {
	CLOUDFLARE_API_TOKEN    string
	CLOUDFLARE_ZONE_ID      string
	CLOUDFLARE_RECORD_NAME  string
	CLOUDFLARE_RECORD_TTL   int
	CLOUDFLARE_RECORD_PROXY bool
	REFRESH_INTERVAL        int
}

func main() {

	// Read config
	config := loadConfig()
	cloudflareClient := cloudflare.NewClient(config.CLOUDFLARE_API_TOKEN, config.CLOUDFLARE_ZONE_ID)

	for {
		if err := job(cloudflareClient, config.CLOUDFLARE_RECORD_NAME, config.CLOUDFLARE_RECORD_TTL); err != nil {
			log.Println(err)
		}
		time.Sleep(time.Duration(config.REFRESH_INTERVAL) * time.Minute)
	}
}

func job(client *cloudflare.Client, recordName string, ttl int) error {

	ipAddress, err := getCurrentIP()
	if err != nil {
		return err
	}

	log.Println("Your ip address is", ipAddress)

	if record, err := client.DnsRecord(recordName); err != nil {
		return err
	} else if record != nil {
		// Record already present, overwrite it with the new address
		log.Printf("DNS record %s found in zone %s", recordName, client.DNSZoneId)
		updated, err := client.UpdateDNSRecord(
			cloudflare.DNSRecordOptions{
				Name:    recordName,
				Value:   ipAddress,
				TTL:     ttl,
				Proxied: false,
			}, record.ID)

		if err != nil {
			return err
		}

		if updated {
			log.Println("DNS record succesfully updated!")
		} else {
			log.Println("DNS record not updated")
		}
	} else {
		// Record do not exists, create a new one
		record_id, err := client.CreateDNSRecord(cloudflare.DNSRecordOptions{
			Name:    recordName,
			Value:   ipAddress,
			TTL:     ttl,
			Proxied: false,
		})
		if err != nil {
			return err
		}

		log.Printf("DNS record succesfully created with id: %s", record_id)
	}

	return nil
}

func loadConfig() *Config {
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

func getCurrentIP() (ip string, err error) {

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
