package main

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

type CloudflareCreateRecordAPIResponse struct {
	Errors   []APIMessage    `json:"errors"`
	Messages []APIMessage    `json:"messages"`
	Success  bool            `json:"success"`
	Result   DDNRecordResult `json:"result"`
}

type CloudflareListRecordAPIResponse struct {
	Errors   []APIMessage      `json:"errors"`
	Messages []APIMessage      `json:"messages"`
	Success  bool              `json:"success"`
	Result   []DDNRecordResult `json:"result"`
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

type DDNRecordResult struct {
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
