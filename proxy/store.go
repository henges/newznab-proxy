package proxy

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (s *Store) GetFeedItemIDs(ctx context.Context, ids []string) (map[string]struct{}, error) {

	rows, err := s.q.GetFeedItemIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	ret := lo.Associate(rows, func(item string) (string, struct{}) {
		return item, struct{}{}
	})
	return ret, nil
}

func (s *Store) InsertFeedItem(ctx context.Context, fi FeedItem) error {

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
		Comments:        nullStr(fi.Comments),
		Description:     nullStr(fi.Description),
		Link:            nullStr(fi.Link),
		NzbUrl:          fi.NZBLink,
		PubDate:         timeToString(fi.PubDate),
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

func (s *Store) SearchForFeedItem(ctx context.Context, search string) ([]FeedItem, error) {

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

	ret := lo.Map(rows, func(item querier.FeedItem, index int) FeedItem {
		guidIsPermalink := false
		if item.GuidIsPermalink == 1 {
			guidIsPermalink = true
		}
		meta, _ := metas[item.ID]

		pubDate, _ := timeFromString(item.PubDate)

		return FeedItem{
			ID:              item.ID,
			IndexerName:     item.IndexerName,
			Title:           item.Title,
			GUID:            item.Guid.String,
			GUIDIsPermalink: guidIsPermalink,
			Link:            item.Link.String,
			Comments:        item.Comments.String,
			PubDate:         pubDate,
			Category:        item.Category.String,
			Description:     item.Description.String,
			NZBLink:         item.NzbUrl,
			Size:            item.Size.Int64,
			Attrs:           meta,
		}
	})
	return ret, nil
}

func timeToString(t time.Time) string {

	return t.Format(time.RFC3339)
}

func timeFromString(s string) (time.Time, error) {

	return time.Parse(time.RFC3339, s)
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

func (s *Store) UpsertSearchCacheEntry(ctx context.Context, entry SearchCacheEntry) error {

	return s.q.UpsertSearchCache(ctx, querier.UpsertSearchCacheParams{
		IndexerName:  entry.IndexerName,
		Query:        entry.Query,
		Categories:   "",
		FirstTried:   entry.FirstTried.Unix(),
		LastTried:    entry.LastTried.Unix(),
		Status:       string(entry.SearchResultStatus),
		ErrorMessage: nullStr(entry.ErrorMessage),
	})
}

func (s *Store) LoadCurrentSearchCacheEntriesForQuery(ctx context.Context, query string, after time.Time) (map[string]SearchCacheEntry, error) {

	rows, err := s.q.LoadCurrentSearchCacheEntriesForQuery(ctx, querier.LoadCurrentSearchCacheEntriesForQueryParams{
		Query:     query,
		LastTried: after.Unix(),
	})
	if err != nil {
		return nil, err
	}
	ret := make(map[string]SearchCacheEntry, len(rows))
	for _, item := range rows {
		ret[item.IndexerName] = SearchCacheEntry{
			IndexerName:        item.IndexerName,
			Query:              item.Query,
			FirstTried:         time.Unix(item.FirstTried, 0),
			LastTried:          time.Unix(item.LastTried, 0),
			SearchResultStatus: SearchResultStatus(item.Status),
			ErrorMessage:       item.ErrorMessage.String,
		}
	}
	return ret, nil
}

func (s *Store) GetNZBDataByID(ctx context.Context, id string) (NZBData, error) {

	row, err := s.q.GetNZBDataByID(ctx, id)
	if err != nil {
		return NZBData{}, err
	}
	ret := NZBData{
		Title:       row.Title,
		IndexerName: row.IndexerName,
		URL:         row.NzbUrl,
	}
	return ret, nil
}
