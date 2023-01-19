package etherscan

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gsrai/go-ionic/config"
)

const NO_CODE = `0x`

func IsContract(address string) bool {
	c := http.Client{Timeout: time.Duration(10) * time.Second}

	req, err := http.NewRequest("GET", config.Get().EtherscanAPI.URL, nil)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("apiKey", config.Get().EtherscanAPI.Token)
	q.Add("module", "proxy")
	q.Add("address", address)
	q.Add("tag", "latest")
	q.Add("action", "eth_getCode")
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

	res := EtherscanAPIResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Panic(err)
	}

	if res.Error.Message != "" {
		return false
	}

	return res.Result != NO_CODE
}
