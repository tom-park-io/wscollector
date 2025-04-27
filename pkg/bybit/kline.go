package bybit

import (
	"strconv"
	"time"

	"wscollector/internal/bybit/memorystore"
)

// ParseKlineList converts Bybit REST API kline data to []Kline.
// It safely skips invalid rows and sets default values for fields not included in REST (e.g., Confirm).
func ParseKlineList(interval string, raw [][]string) ([]memorystore.Kline, error) {
	var out []memorystore.Kline

	for _, row := range raw {
		if len(row) < 7 {
			continue // skip incomplete row
		}

		start, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue
		}
		open, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}
		high, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			continue
		}
		low, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			continue
		}
		closeVal, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			continue
		}
		volume, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			continue
		}
		turnover, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			continue
		}

		out = append(out, memorystore.Kline{
			Start:     start,
			End:       time.UnixMilli(start).Add(time.Minute).UnixMilli(), // fixed to 1m
			Interval:  interval,
			Open:      strconv.FormatFloat(open, 'f', -1, 64),
			High:      strconv.FormatFloat(high, 'f', -1, 64),
			Low:       strconv.FormatFloat(low, 'f', -1, 64),
			Close:     strconv.FormatFloat(closeVal, 'f', -1, 64),
			Volume:    strconv.FormatFloat(volume, 'f', -1, 64),
			Turnover:  strconv.FormatFloat(turnover, 'f', -1, 64),
			Confirm:   true,
			Timestamp: time.Now().UnixMilli(), // Time of ingestion // or use Start.Add(...) as approximation
		})
	}
	return out, nil
}
