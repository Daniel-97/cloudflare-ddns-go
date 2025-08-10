package main

import (
	"log"
	"time"

	"cloudflare-ddns-go/cloudflare"
)

func main() {

	// Read config
	config := loadConfig()
	cloudflareClient := cloudflare.NewClient(config.CLOUDFLARE_API_TOKEN, config.CLOUDFLARE_ZONE_ID)

	for {
		if err := job(cloudflareClient, config.CLOUDFLARE_RECORD_NAME, config.CLOUDFLARE_RECORD_TTL); err != nil {
			log.Printf("Unexpected error in job: %s", err)
		}
		time.Sleep(time.Duration(config.REFRESH_INTERVAL) * time.Minute)
	}
}

func job(client *cloudflare.Client, recordName string, ttl int) error {

	ipAddress, err := cloudflare.CurrentIP()
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
