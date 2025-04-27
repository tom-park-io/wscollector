package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"wscollector/internal/bybit/memorystore"

	"gorm.io/gorm/clause"
)

func (p *PostgresClient) InsertKline(ctx context.Context, record *KlineRecord) error {
	tx := p.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "interval"},
			{Name: "start"},
			{Name: "confirm"},
		},
		DoNothing: true,
	}).Create(record)

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf(
			"duplicate kline skipped: symbol=%s interval=%s start=%s confirm=%t",
			record.Symbol,
			record.Interval,
			record.Start.Format(time.RFC3339),
			record.Confirm,
		)
	}

	return nil
}

// example methods
func (p *PostgresClient) GetKline(ctx context.Context, symbol, interval string, start time.Time) (*KlineRecord, error) {
	var kline KlineRecord
	err := p.DB.WithContext(ctx).
		Where("symbol = ? AND interval = ? AND start = ?", symbol, interval, start).
		First(&kline).Error

	if err != nil {
		return nil, err
	}
	return &kline, nil
}

func (p *PostgresClient) UpdateKlineConfirm(ctx context.Context, id uint, confirm bool) error {
	return p.DB.WithContext(ctx).
		Model(&KlineRecord{}).
		Where("id = ?", id).
		Update("confirm", confirm).Error
}

func (p *PostgresClient) DeleteOldKlines(ctx context.Context, before time.Time) error {
	return p.DB.WithContext(ctx).
		Where("start < ?", before).
		Delete(&KlineRecord{}).Error
}

// ToKlineRecord converts a Kline and symbol into a KlineRecord for DB insertion.
func ToKlineRecord(symbol string, k memorystore.Kline) (*KlineRecord, error) {
	open, err := strconv.ParseFloat(k.Open, 64)
	if err != nil {
		return nil, err
	}
	closePrice, err := strconv.ParseFloat(k.Close, 64)
	if err != nil {
		return nil, err
	}
	high, err := strconv.ParseFloat(k.High, 64)
	if err != nil {
		return nil, err
	}
	low, err := strconv.ParseFloat(k.Low, 64)
	if err != nil {
		return nil, err
	}
	volume, err := strconv.ParseFloat(k.Volume, 64)
	if err != nil {
		return nil, err
	}
	turnover, err := strconv.ParseFloat(k.Turnover, 64)
	if err != nil {
		return nil, err
	}

	return &KlineRecord{
		Symbol: symbol,
		// TODO: define interval symbols
		Interval:  fmt.Sprintf("%sm", k.Interval),
		Start:     time.UnixMilli(k.Start),
		End:       time.UnixMilli(k.End),
		Open:      open,
		Close:     closePrice,
		High:      high,
		Low:       low,
		Volume:    volume,
		Turnover:  turnover,
		Confirm:   k.Confirm,
		Timestamp: time.UnixMilli(k.Timestamp),
	}, nil
}
