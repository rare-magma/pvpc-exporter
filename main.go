package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
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

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
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
