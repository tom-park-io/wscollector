package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wscollector/internal/bybit/memorystore"
)

type RESTClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRESTClient(baseURL string, timeout time.Duration) *RESTClient {
	return &RESTClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *RESTClient) HTTPClient() *http.Client {
	return c.httpClient
}

// GetUSDTAltcoinSymbols fetches linear symbols with quoteCoin = USDT (altcoins).
func (c *RESTClient) GetUSDTAltcoinSymbols(ctx context.Context) ([]string, error) {
	endpoint := c.baseURL + "/v5/market/instruments-info?category=linear&limit=1000"

	// Construct the GET request with context for timeout/cancel support
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Execute the HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bybit error: %s", body)
	}

	var rawResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Decode result into InstrumentListResponse
	var result InstrumentListResponse
	if err := json.Unmarshal(rawResp.Result, &result); err != nil {
		return nil, fmt.Errorf("decode result: %w", err)
	}

	// Collect USDT-based altcoins
	seen := map[string]bool{}
	var baseCoins []string
	for _, symbol := range result.List {
		if symbol.QuoteCoin == "USDT" && !seen[symbol.BaseCoin] {
			// baseCoins = append(baseCoins, symbol.BaseCoin)
			baseCoins = append(baseCoins, symbol.Symbol)
			seen[symbol.BaseCoin] = true
		}
	}

	return baseCoins, nil
}

func (c *RESTClient) GetKlines(ctx context.Context, category, symbol, interval string,
	start, end time.Time) ([]memorystore.Kline, error) {
	endpoint := fmt.Sprintf(
		"%s/v5/market/kline?category=%s&symbol=%s&interval=%s&start=%d&end=%d",
		c.baseURL,
		category,
		symbol,
		interval,
		start.UnixMilli(),
		end.UnixMilli(),
	)

	// Construct the GET request with context for timeout/cancel support
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute the HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bybit error: %s", body)
	}

	var rawResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Decode result into KlinesResponse
	var result KlinesResponse
	if err := json.Unmarshal(rawResp.Result, &result); err != nil {
		return nil, fmt.Errorf("decode result: %w", err)
	}

	// TODO: define interval
	klines, err := ParseKlineList(fmt.Sprintf("%s", interval), result.List)
	if err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}

	return klines, nil
}
