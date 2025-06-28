package proxy

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/henges/newznab-proxy/proxy/model"
	"github.com/henges/newznab-proxy/proxy/querier"
	"github.com/samber/lo"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
	q  *querier.Queries
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
	return &Store{db: db, q: querier.New(db)}, nil
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (s *Store) InsertFeedItem(ctx context.Context, fi model.FeedItem) error {

	isPerma := int64(0)
	if fi.GUIDIsPermalink {
		isPerma = 1
	}

	err := s.q.InsertFeedItem(ctx, querier.InsertFeedItemParams{
		ID:              fi.ID,
		IndexerName:     fi.IndexerName,
		Title:           fi.Title,
		Guid:            nullStr(fi.GUID),
		GuidIsPermalink: isPerma,
		Link:            nullStr(fi.Link),
		NzbUrl:          fi.NZBLink,
		PubDate:         fi.PubDate,
		Size: sql.NullInt64{
			Int64: fi.Size,
			Valid: true,
		},
		Category: nullStr(fi.Category),
		Source:   string(fi.Source),
	})
	if err != nil {
		return err
	}
	for k, v := range fi.Attrs {
		err = s.q.InsertFeedItemMeta(ctx, querier.InsertFeedItemMetaParams{
			FeedItemID: fi.ID,
			Name:       k,
			Value:      v,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SearchForFeedItem(ctx context.Context, search string) ([]model.FeedItem, error) {

	rows, err := s.q.SearchForFeedItem(ctx, search)
	if err != nil {
		return nil, err
	}
	ids := lo.Map(rows, func(item querier.FeedItem, index int) string {
		return item.ID
	})
	metas, err := s.GetFeedItemMetas(ctx, ids)
	if err != nil {
		return nil, err
	}

	ret := lo.Map(rows, func(item querier.FeedItem, index int) model.FeedItem {
		guidIsPermalink := false
		if item.GuidIsPermalink == 1 {
			guidIsPermalink = true
		}
		meta, _ := metas[item.ID]

		return model.FeedItem{
			ID:              item.ID,
			IndexerName:     item.IndexerName,
			Title:           item.Title,
			GUID:            item.Guid.String,
			GUIDIsPermalink: guidIsPermalink,
			Link:            item.Link.String,
			Comments:        item.Comments.String,
			PubDate:         item.PubDate,
			Category:        item.Category.String,
			Description:     item.Description.String,
			NZBLink:         item.NzbUrl,
			Size:            item.Size.Int64,
			Attrs:           meta,
		}
	})
	return ret, nil
}

func (s *Store) GetFeedItemMetas(ctx context.Context, ids []string) (map[string]map[string]string, error) {

	metas, err := s.q.GetFeedItemMetas(ctx, ids)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]map[string]string)
	for _, meta := range metas {
		existing, ok := ret[meta.FeedItemID]
		if !ok {
			existing = make(map[string]string)
			ret[meta.FeedItemID] = existing
		}
		existing[meta.Name] = meta.Value
	}
	return ret, nil
}
