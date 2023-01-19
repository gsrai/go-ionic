package etherscan

type EtherscanError struct {
	Message string `json:"message"`
}

type EtherscanAPIResponse struct {
	Result string         `json:"result"`
	Error  EtherscanError `json:"error"`
}
