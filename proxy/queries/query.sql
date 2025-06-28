-- name: InsertFeedItem :exec
INSERT INTO feed_items (id, indexer_name, title, guid, guid_is_permalink, link, nzb_url, pub_date, size, category, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: SearchForFeedItem :exec
SELECT feed_items.* FROM feed_items
JOIN feed_items_fts5 f on feed_items.id = f.id
WHERE f.title MATCH ?
ORDER BY f.rank;
