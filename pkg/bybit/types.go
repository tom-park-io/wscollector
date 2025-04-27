package bybit

import "encoding/json"

// BybitResponse represents a generic response from Bybit's V5 REST API.
// This structure covers the standard response envelope used across all endpoints.
type BybitResponse struct {
	RetCode    int                    `json:"retCode"`    // 0 means success; non-zero indicates an error code
	RetMsg     string                 `json:"retMsg"`     // Human-readable message describing the result or error
	Result     json.RawMessage        `json:"result"`     // Delay decoding // Main response payload (varies per endpoint)
	RetExtInfo map[string]interface{} `json:"retExtInfo"` // Optional extra info (e.g. rate limits, error hints)
	Time       int64                  `json:"time"`       // Server timestamp (in milliseconds since epoch)
}

type InstrumentListResponse struct {
	Category       string `json:"category"` // e.g., "linear", "spot"
	NextPageCursor string `json:"nextPageCursor"`
	List           []struct {
		Symbol    string `json:"symbol"`    // e.g., "BTCUSDT"
		BaseCoin  string `json:"baseCoin"`  // e.g., "BTC"
		QuoteCoin string `json:"quoteCoin"` // e.g., "USDT"
		// ... extra
	} `json:"list"`
}

type KlinesResponse struct {
	Category       string     `json:"category"` // e.g., "linear", "spot"
	NextPageCursor string     `json:"nextPageCursor"`
	List           [][]string `json:"list"`
}
