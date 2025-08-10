package main

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	CLOUDFLARE_API_TOKEN    string
	CLOUDFLARE_ZONE_ID      string
	CLOUDFLARE_RECORD_NAME  string
	CLOUDFLARE_RECORD_TTL   int
	CLOUDFLARE_RECORD_PROXY bool
	REFRESH_INTERVAL        int
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
