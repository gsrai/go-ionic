package covalent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gsrai/go-ionic/internal/types"
	"github.com/gsrai/go-ionic/internal/utils"
)

const API_BASE_URL = "https://api.covalenthq.com/v1"
const API_KEY = "ckey_0f89a2f9110f48e0837ee6770c9"

func GetBlockHeights(chainId types.ChainID, start, end time.Time) CovalentAPIResponse[BlockHeights] {
	url := fmt.Sprintf("%v/%v/block_v2/%v/%v/", API_BASE_URL, chainId, utils.ToISOString(start), utils.ToISOString(end))
	log.Printf("getting block height, URL: %v", url)

	c := http.Client{Timeout: time.Duration(1) * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Add("Accept", `application/json`)

	q := req.URL.Query()
	q.Add("key", API_KEY)
	q.Add("page-size", "10")
	q.Add("page-number", "0")
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

	res := CovalentAPIResponse[BlockHeights]{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Panic(err)
	}

	return res
}
