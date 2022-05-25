package etherscan

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gsrai/go-ionic/internal/types"
)

const API_BASE_URL = "https://api.etherscan.io/api"
const API_KEY = "1CAF88PW5CPEJ7I6GD43ZCUG3YGH6HAUE1"
const NO_CODE = `0x`

func IsContract(address types.Address) bool {
	c := http.Client{Timeout: time.Duration(10) * time.Second}

	req, err := http.NewRequest("GET", API_BASE_URL, nil)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("apiKey", API_KEY)
	q.Add("module", "proxy")
	q.Add("address", string(address))
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

	if res.Data.Error.Message != "" {
		return false
	}

	return res.Data.Result != NO_CODE
}
