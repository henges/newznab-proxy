-- name: InsertFeedItem :one
INSERT INTO feed_items (uuid, indexer_name, title, guid, guid_is_permalink, link, nzb_url, pub_date, size, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: GetFeedItemUUIDs :many
SELECT uuid FROM feed_items WHERE uuid IN (sqlc.slice(ids));

-- name: InsertFeedItemMeta :exec
INSERT INTO feed_item_meta (feed_item_id, name, value) VALUES (?, ?, ?);

-- name: SearchForFeedItem :many
SELECT feed_items.* FROM feed_items
JOIN feed_items_fts5 f on feed_items.id = f.id
WHERE f.title MATCH ?
ORDER BY f.rank;

-- name: GetFeedItemMetas :many
SELECT * FROM feed_item_meta WHERE feed_item_id IN (sqlc.slice(ids));

-- name: LoadCurrentSearchCacheEntriesForQuery :many
SELECT * FROM search_cache
WHERE query = ? and last_tried >= ?;

-- name: LoadSearchCacheEntriesForQuery :many
SELECT * FROM search_cache
WHERE query = ?;

-- name: UpsertSearchCache :exec
INSERT INTO search_cache (indexer_name,
                          query,
                          categories,
                          first_tried,
                          last_tried,
                          status,
                          error_message)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(indexer_name, query) DO UPDATE SET last_tried    = excluded.last_tried,
                                               status        = excluded.status,
                                               error_message = excluded.error_message;

-- name: GetNZBDataByUUID :one
SELECT title, indexer_name, nzb_url FROM feed_items WHERE uuid = ? LIMIT 1;
