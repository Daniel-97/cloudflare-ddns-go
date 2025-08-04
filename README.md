# cloudflare-ddns-go
A zero dependency go client used to update DNS entries on Cloudflare accounts based on current ip address.

## Prerequisites

- A registered domain
- A cloudflare account
- The Go compiler
- Some spare time

## Supported IP version

Both ip version are supported:

- Ipv4 -> A
- Ipv6 -> AAAA

## Enviroment variables

| Variable Name       | Description                              | Default                   |
| ------------------| ---------------------------------------- | ------------------------- |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API key | N/A (required)                      |
| `CLOUDFLARE_ZONE_ID`         | Your cloudflare dns record zone id            | N/A (required)|
| `CLOUDFLARE_RECORD_NAME` | Record name (e.g. subdomain.mydomain.com)               | N/A (required)|
| `CLOUDFLARE_RECORD_TTL`       | DNS record TTL (Time to live)         |3600 (1h)|
| `CLOUDFLARE_RECORD_PROXY`       | DNS record proxy         |false|
| `REFRESH_INTERVAL`       | DNS record update interval        | 5 (min) |

## Docker


### Build
```bash
docker build -t cloudflare-ddns-go .
```

### Run
```bash
docker run --rm \
  -e CLOUDFLARE_API_TOKEN=your_cloudflare_api_key \
  -e CLOUDFLARE_ZONE_ID=your-cloudflare_zone_id \
  -e CLOUDFLARE_RECORD_NAME=subdomain.yourdomain.com \
  -e CLOUDFLARE_RECORD_TTL=3600 \
  -e CLOUDFLARE_RECORD_PROXY=false \
  -e REFRESH_INTERVAL=5 \
  cloudflare-ddns-go
```

## Disclaimer
I am not a professional golang developer, use this tool with caution

