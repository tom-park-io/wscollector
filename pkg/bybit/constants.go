package bybit

import "fmt"

// KlineInterval is the interval type used for API requests
type KlineInterval string

// KlineIntervalMeta holds API value and DB value for a Kline interval
type KlineIntervalMeta struct {
	APIValue string
	DBValue  string
	Minutes  int
}

const (
	Interval1Min    KlineInterval = "1"
	Interval3Min    KlineInterval = "3"
	Interval5Min    KlineInterval = "5"
	Interval15Min   KlineInterval = "15"
	Interval30Min   KlineInterval = "30"
	Interval60Min   KlineInterval = "60"
	Interval120Min  KlineInterval = "120"
	Interval240Min  KlineInterval = "240"
	Interval360Min  KlineInterval = "360"
	Interval720Min  KlineInterval = "720"
	IntervalDaily   KlineInterval = "D"
	IntervalWeekly  KlineInterval = "W"
	IntervalMonthly KlineInterval = "M"
)

// validKlineIntervals maps KlineInterval to its API and DB representations
var validKlineIntervals = map[KlineInterval]KlineIntervalMeta{
	Interval1Min:    {APIValue: "1", DBValue: "1m", Minutes: 1},
	Interval3Min:    {APIValue: "3", DBValue: "3m", Minutes: 3},
	Interval5Min:    {APIValue: "5", DBValue: "5m", Minutes: 5},
	Interval15Min:   {APIValue: "15", DBValue: "15m", Minutes: 15},
	Interval30Min:   {APIValue: "30", DBValue: "30m", Minutes: 30},
	Interval60Min:   {APIValue: "60", DBValue: "1h", Minutes: 60},
	Interval120Min:  {APIValue: "120", DBValue: "2h", Minutes: 120},
	Interval240Min:  {APIValue: "240", DBValue: "4h", Minutes: 240},
	Interval360Min:  {APIValue: "360", DBValue: "6h", Minutes: 360},
	Interval720Min:  {APIValue: "720", DBValue: "12h", Minutes: 720},
	IntervalDaily:   {APIValue: "D", DBValue: "1d", Minutes: 1440},  // 24*60
	IntervalWeekly:  {APIValue: "W", DBValue: "1w", Minutes: 10080}, // 7*24*60
	IntervalMonthly: {APIValue: "M", DBValue: "1M", Minutes: 43200}, // 30*24*60 // TODO: FIX - 30 days assumption, handle actual month duration
}

// IsValid checks if the KlineInterval is a valid predefined interval
func (k KlineInterval) IsValid() bool {
	_, ok := validKlineIntervals[k]
	return ok
}

// ParseKlineInterval parses a string into a valid KlineIntervalMeta
func ParseKlineInterval(s string) (KlineIntervalMeta, error) {
	interval := KlineInterval(s)
	meta, ok := validKlineIntervals[interval]
	if !ok {
		return KlineIntervalMeta{}, fmt.Errorf("invalid KlineInterval: %s", s)
	}
	return meta, nil
}
