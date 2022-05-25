package etherscan

type EtherscanError struct {
	Message string `json:"message"`
}

type EtherscanData struct {
	Result string         `json:"result"`
	Error  EtherscanError `json:"error"`
}

type EtherscanAPIResponse struct {
	Data EtherscanData `json:"data"`
}
