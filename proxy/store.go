package proxy

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/henges/newznab-proxy/newznab"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, path string) (*Store, error) {

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%v", path))
	if err != nil {
		return nil, err
	}
	err = migrateDB(ctx, db)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func SearchForFeedItem(ctx context.Context, search string) ([]newznab.Item, error) {

}
