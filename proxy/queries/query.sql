-- name: InsertFeedItem :exec
INSERT INTO feed_items (id, indexer_name, title, guid, guid_is_permalink, link, nzb_url, pub_date, size, category, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertFeedItemMeta :exec
INSERT INTO feed_item_meta (feed_item_id, name, value) VALUES (?, ?, ?);

-- name: SearchForFeedItem :many
SELECT feed_items.* FROM feed_items
JOIN feed_items_fts5 f on feed_items.id = f.id
WHERE f.title MATCH ?
ORDER BY f.rank;

-- name: GetFeedItemMetas :many
SELECT * FROM feed_item_meta WHERE feed_item_id IN (sqlc.slice(ids));

-- name: