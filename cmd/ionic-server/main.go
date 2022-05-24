package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gsrai/go-ionic/internal/types"
	"github.com/gsrai/go-ionic/internal/utils"
	"github.com/gsrai/go-ionic/internal/utils/csv"
)

const HOST = "127.0.0.1"
const PORT = "8080"
const INPUT_FILE_PATH = "tmp/input.csv"
const API_BASE_URL = "https://api.covalenthq.com/v1"
const API_KEY = "ckey_0f89a2f9110f48e0837ee6770c9"

func main() {
	addr := HOST + ":" + PORT

	http.HandleFunc("/input/load", loadData)
	http.HandleFunc("/block/heights", getRecentBlockHeight)

	log.Print("Running server on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getRecentBlockHeight(w http.ResponseWriter, req *http.Request) {
	start := time.Date(2021, 11, 10, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	res := getBlockHeights(types.ETH, start, end)
	fmt.Fprintf(w, "block heights: %v\n", res)
}

func loadData(w http.ResponseWriter, req *http.Request) {
	mapperFunc := func(csvRow []string) (types.InputCSVRecord, error) {
		rate, err := strconv.ParseFloat(csvRow[5], 64)
		if err != nil {
			return types.InputCSVRecord{}, err
		}

		from, err := utils.ParseDateTime(csvRow[3])
		if err != nil {
			return types.InputCSVRecord{}, err
		}

		to, err := utils.ParseDateTime(csvRow[4])
		if err != nil {
			return types.InputCSVRecord{}, err
		}

		return types.InputCSVRecord{
			CoinName:     csvRow[0],
			ContractAddr: csvRow[1],
			From:         from,
			To:           to,
			Network:      csvRow[2],
			Rate:         rate,
		}, nil
	}

	data := csv.ReadAndParse(INPUT_FILE_PATH, mapperFunc)

	for idx, datum := range data {
		fmt.Fprintf(w, "row %d: %v\n", idx+1, datum)
	}
}

type Block struct {
	SignedAt time.Time `json:"signed_at"`
	Height   int
}

type BlockHeights struct {
	UpdatedAt  time.Time          `json:"updated_at"`
	Items      []Block            `json:"items"`
	Pagination CovalentPagination `json:"pagination"`
}

type CovalentPagination struct {
	HasMore    bool `json:"has_more"`
	PageNumber int  `json:"page_number"`
	PageSize   int  `json:"page_size"`
}

type CovalentAPIResponse[T any] struct {
	Data         T      `json:"data"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
	ErrorCode    int    `json:"error_code"`
}

func getBlockHeights(chainId types.ChainID, start, end time.Time) CovalentAPIResponse[BlockHeights] {
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
