package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	APIToken  string
	DNSZoneId string
}

type APIRequest struct {
	Name    string `json:"name"`
	TTL     int    `json:"ttl"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

type CreateRecordAPIResponse struct {
	Errors   []APIMessage     `json:"errors"`
	Messages []APIMessage     `json:"messages"`
	Success  bool             `json:"success"`
	Result   DDNSRecordResult `json:"result"`
}

type ListRecordAPIResponse struct {
	Errors   []APIMessage       `json:"errors"`
	Messages []APIMessage       `json:"messages"`
	Success  bool               `json:"success"`
	Result   []DDNSRecordResult `json:"result"`
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

type DDNSRecordResult struct {
	Name     string   `json:"name"`
	TTL      int      `json:"ttl"`
	Type     string   `json:"type"`
	Comment  string   `json:"comment"`
	Content  string   `json:"content"`
	Proxied  bool     `json:"proxied"`
	Settings settings `json:"settings"`
	Tags     []string `json:"tags"`
	ID       string   `json:"id"`
}

type settings struct {
	IPv4Only bool `json:"ipv4_only"`
	IPv6Only bool `json:"ipv6_only"`
}

func NewClient(APIToken string, DNSZoneId string) *Client {
	return &Client{APIToken: APIToken, DNSZoneId: DNSZoneId}
}

func setHeader(req *http.Request, apiToken string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
}

func (c Client) DnsRecord(recordName string) (*DDNSRecordResult, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", c.DNSZoneId)
	log.Println("Searching dns record", recordName)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	setHeader(req, c.APIToken)

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	cloudflare_response, err := ParseResponse[ListRecordAPIResponse](res)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d (%s): %s", res.StatusCode, url, cloudflare_response.Errors[0].Message)
	}

	log.Printf("Found %d dns records", len(cloudflare_response.Result))
	for _, record := range cloudflare_response.Result {
		if record.Name == recordName {
			return &record, nil
		}
	}

	return nil, nil

}

func (c Client) UpdateDNSRecord(recordName string, recordTTL int, recordProxy bool, record_id string, address string) (bool, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", c.DNSZoneId, record_id)

	req_body := DDNSRecordResult{
		Name:    recordName,
		TTL:     recordTTL,
		Type:    dnsRecordType(address),
		Comment: fmt.Sprintf("cloudflare-ddns-go (%s)", time.Now().Format(time.RFC3339)),
		Content: address,
		Proxied: recordProxy,
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

	setHeader(req, c.APIToken)

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return false, err
	}

	cloudflare_response, err := ParseResponse[CreateRecordAPIResponse](res)
	if err != nil {
		return false, err
	}

	if res.StatusCode == http.StatusOK {
		return true, nil

	} else {
		return false, fmt.Errorf("HTTP error %d (%s): %s", res.StatusCode, url, cloudflare_response.Errors[0].Message)
	}

}

func (c Client) CreateDNSRecord(recordName string, recordTTL int, recordProxy bool, address string) (record_id string, err error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", c.DNSZoneId)

	req_body := DDNSRecordResult{
		Name:    recordName,
		TTL:     recordTTL,
		Type:    dnsRecordType(address),
		Comment: fmt.Sprintf("cloudflare-ddns-go (%s)", time.Now().Format(time.RFC3339)),
		Content: address,
		Proxied: recordProxy,
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

	setHeader(req, c.APIToken)

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	cloudflare_response, err := ParseResponse[CreateRecordAPIResponse](res)
	if err != nil {
		return "", err
	}

	if res.StatusCode == http.StatusOK {
		return cloudflare_response.Result.ID, nil

	} else {
		return "", fmt.Errorf("HTTP error %d (%s): %s", res.StatusCode, url, cloudflare_response.Errors[0].Message)
	}

}

func ParseResponse[T any](response *http.Response) (json_body T, err error) {

	var result T
	body, err := io.ReadAll(response.Body)

	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}

func dnsRecordType(address string) string {
	if strings.Count(address, ":") >= 2 {
		return "AAAA"
	} else {
		return "A"
	}
}
