package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type PriceData struct {
	Value    float64 `json:"value"`
	Datetime string  `json:"datetime"`
}

type PVPCAttribute struct {
	Values []PriceData `json:"values"`
}

type PVPCIncluded struct {
	Id         string        `json:"id"`
	Attributes PVPCAttribute `json:"attributes"`
}

type PVPCJson struct {
	Included []PVPCIncluded `json:"included"`
}

type Config struct {
	Bucket           string `json:"Bucket"`
	InfluxDBHost     string `json:"InfluxDBHost"`
	InfluxDBApiToken string `json:"InfluxDBApiToken"`
	Org              string `json:"Org"`
}

type retryableTransport struct {
	transport             http.RoundTripper
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
}

const retryCount = 3

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}
	if resp == nil {
		return true
	}
	switch resp.StatusCode {
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (t *retryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	resp, err := t.transport.RoundTrip(req)
	retries := 0
	for shouldRetry(err, resp) && retries < retryCount {
		backoff := time.Duration(math.Pow(2, float64(retries))) * time.Second
		time.Sleep(backoff)
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		if resp != nil {
			log.Printf("Previous request failed with %s", resp.Status)
		}
		log.Printf("Retry %d of request to: %s", retries+1, req.URL)
		resp, err = t.transport.RoundTrip(req)
		retries++
	}
	return resp, err
}

func main() {
	confFilePath := "pvpc_exporter.json"
	confData, err := os.Open(confFilePath)
	if err != nil {
		log.Fatalln("Error reading config file: ", err)
	}
	defer confData.Close()
	var config Config
	err = json.NewDecoder(confData).Decode(&config)
	if err != nil {
		log.Fatalln("Error reading configuration: ", err)
	}
	if config.Bucket == "" {
		log.Fatalln("Bucket is required")
	}
	if config.InfluxDBHost == "" {
		log.Fatalln("InfluxDBHost is required")
	}
	if config.InfluxDBApiToken == "" {
		log.Fatalln("InfluxDBApiToken is required")
	}
	if config.Org == "" {
		log.Fatalln("Org is required")
	}

	var days int
	flag.IntVar(&days, "days", 0, "Number of days in the past to fetch")
	flag.Parse()

	transport := &retryableTransport{
		transport:             &http.Transport{},
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	date := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	const apiUrl = "https://apidatos.ree.es/es/datos/mercados/precios-mercados-tiempo-real"
	pvpcURL := fmt.Sprintf(apiUrl+"?start_date=%sT00:00&end_date=%sT23:59&time_trunc=hour", date, date)
	req, _ := http.NewRequest("GET", pvpcURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error fetching data: ", err)
	}
	defer resp.Body.Close()
	getStatusOK := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest
	if !getStatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln("Error reading data: ", err)
		}
		log.Fatalln("Error fetching PVPC data: ", string(resp.Status), string(body))
	}
	pvpcData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Error reading data: ", err)
	}
	var pvpc PVPCJson
	err = json.Unmarshal(pvpcData, &pvpc)
	if err != nil {
		log.Fatalln("Error unmarshalling data: ", err)
	}

	payload := bytes.Buffer{}
	for _, inc := range pvpc.Included {
		if inc.Id == "1001" {
			for _, value := range inc.Attributes.Values {
				timestamp, err := time.Parse(time.RFC3339, value.Datetime)
				if err != nil {
					log.Fatalln("Error parsing timestamp: ", err)
				}
				influxLine := fmt.Sprintf("pvpc_price price=%.2f %v\n", value.Value, timestamp.Unix())
				payload.WriteString(influxLine)
			}
		}
	}

	if len(payload.Bytes()) == 0 {
		log.Fatalln("No data to send")
	}
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(payload.Bytes())
	err = w.Close()
	if err != nil {
		log.Fatalln("Error compressing data: ", err)
	}
	url := fmt.Sprintf("https://%s/api/v2/write?precision=s&org=%s&bucket=%s", config.InfluxDBHost, config.Org, config.Bucket)
	post, _ := http.NewRequest("POST", url, &buf)
	post.Header.Set("Accept", "application/json")
	post.Header.Set("Authorization", "Token "+config.InfluxDBApiToken)
	post.Header.Set("Content-Encoding", "gzip")
	post.Header.Set("Content-Type", "text/plain; charset=utf-8")
	postResp, err := client.Do(post)
	if err != nil {
		log.Fatalln("Error sending data: ", err)
	}
	defer postResp.Body.Close()
	statusOK := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	if !statusOK {
		body, err := io.ReadAll(postResp.Body)
		if err != nil {
			log.Fatalln("Error reading data: ", err)
		}
		log.Fatalln("Error sending data: ", postResp.Status, string(body))
	}
}
