package storage

type Trade struct {
	Symbol string
	Price  float64
	Volume float64
}

type Snapshot struct {
	Symbol    string
	State     string
	Timestamp int64
}
