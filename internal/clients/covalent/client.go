package covalent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gsrai/go-ionic/config"
	"github.com/gsrai/go-ionic/internal/types"
	"github.com/gsrai/go-ionic/internal/utils"
)

func GetBlockHeights(chainId types.ChainID, start, end time.Time) CovalentAPIResponse[Block] {
	url := fmt.Sprintf("%v/%v/block_v2/%v/%v/", config.Get().CovalentAPI.URL, chainId, utils.ToISOString(start), utils.ToISOString(end))
	log.Printf("getting block height, URL: %v", url)

	headers := map[string]string{"Accept": `application/json`}
	params := map[string]string{"page-size": "10", "page-number": "0"}

	return GetRequest[Block](url, headers, params)
}

func GetLogEvents(contractAddr string, startBlock int, endBlock int, chainId types.ChainID) CovalentAPIResponse[LogEvent] {
	url := fmt.Sprintf("%v/%v/events/address/%v/", config.Get().CovalentAPI.URL, chainId, contractAddr)
	log.Printf("getting log events, URL: %v", url)

	headers := map[string]string{"Accept": `application/json`}
	params := map[string]string{
		"quote-currency": "USD",
		"format":         "JSON",
		"starting-block": strconv.Itoa(startBlock),
		"ending-block":   strconv.Itoa(endBlock),
		"page-size":      strconv.Itoa(5000),
	}

	return paginatedGetRequest[LogEvent](url, headers, params)
}

func GetRequest[T APIResponse](url string, headers map[string]string, params map[string]string) CovalentAPIResponse[T] {
	c := http.Client{Timeout: time.Duration(10) * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	q := req.URL.Query()
	q.Add("key", config.Get().CovalentAPI.Token)
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	log.Println(req.URL.String())

	resp, err := c.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	res := CovalentAPIResponse[T]{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Panic(err)
	}

	return res
}

func paginatedGetRequest[T APIResponse](url string, headers map[string]string, params map[string]string) CovalentAPIResponse[T] {
	var items []T
	page := 0

	for {
		// log.Printf("fetching page %d\n", page)
		params["page-number"] = strconv.Itoa(page)
		res := GetRequest[T](url, headers, params)
		if len(res.Data.Items) == 0 {
			return CovalentAPIResponse[T]{
				Data: CovalentDataBody[T]{
					UpdatedAt:  res.Data.UpdatedAt,
					Items:      items,
					Pagination: res.Data.Pagination,
				},
			}
		}
		items = append(items, res.Data.Items...)
		page++
	}
}
