package storage

type Store interface {
	SaveTrade(trade Trade) error
	SaveSnapshot(snapshot Snapshot) error
}
