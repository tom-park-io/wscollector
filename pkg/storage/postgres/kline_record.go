package postgres

import "time"

// KlineRecord represents a finalized candlestick stored in the database.
type KlineRecord struct {
	ID uint `gorm:"primaryKey"`

	// unique index
	Symbol   string    `gorm:"type:text;not null;index:idx_kline_symbol;index:idx_symbol_interval_start_confirm,unique"`
	Interval string    `gorm:"type:varchar(10);not null;index:idx_symbol_interval_start_confirm,unique"`
	Start    time.Time `gorm:"not null;index:idx_symbol_interval_start_confirm,unique"`
	Confirm  bool      `gorm:"not null;index:idx_symbol_interval_start_confirm,unique"`

	End time.Time `gorm:"not null"`

	Open  float64 `gorm:"type:numeric;not null"`
	Close float64 `gorm:"type:numeric;not null"`
	High  float64 `gorm:"type:numeric;not null"`
	Low   float64 `gorm:"type:numeric;not null"`

	Volume   float64 `gorm:"type:numeric;not null"`
	Turnover float64 `gorm:"type:numeric;not null"`

	Timestamp time.Time `gorm:"not null;index:idx_kline_timestamp"`

	RecordedAt time.Time `gorm:"autoCreateTime"`
}

// TableName overrides the default table name for GORM.
func (KlineRecord) TableName() string {
	return "kline_record"
}
