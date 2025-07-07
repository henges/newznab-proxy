-- Recreate feed_items and dependent tables so that the int PK is used
DROP TRIGGER feed_items_fts5_populate;
DROP TABLE feed_items_fts5;

CREATE TABLE feed_items_new
(
    id                INTEGER  PRIMARY KEY AUTOINCREMENT,
    uuid              TEXT     NOT NULL UNIQUE,  -- indexer_name:GUID
    indexer_name      TEXT     NOT NULL,
    title             TEXT     NOT NULL,
    guid              TEXT,              -- original feed GUID (may not be unique across indexers)
    guid_is_permalink int      not null default 1,
    link              TEXT,
    nzb_url           TEXT     NOT NULL,
    pub_date          TEXT     NOT NULL,
    size              INTEGER,           -- in bytes
    source            TEXT     NOT NULL, -- 'rss' or 'search'
    created_at        DATETIME          DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO feed_items_new (uuid, indexer_name, title, guid, guid_is_permalink, link, nzb_url, pub_date, size, source, created_at)
SELECT id, indexer_name, title, guid, guid_is_permalink, link, nzb_url, pub_date, size, source, created_at FROM feed_items;

CREATE TABLE feed_item_meta_new
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_item_id INT NOT NULL REFERENCES feed_items_new (id),
    name         TEXT NOT NULL,
    value        TEXT NOT NULL
);

INSERT INTO feed_item_meta_new (id, feed_item_id, name, value)
SELECT fim.id, fin.id, name, value FROM feed_item_meta fim
JOIN feed_items_new fin on fin.uuid = fim.feed_item_id;

DROP TABLE feed_item_meta;
DROP TABLE feed_items;
ALTER TABLE feed_item_meta_new RENAME TO feed_item_meta;
ALTER TABLE feed_items_new RENAME TO feed_items;

CREATE VIRTUAL TABLE feed_items_fts5 USING fts5
(
    title,
    id UNINDEXED
);

CREATE TRIGGER feed_items_fts5_populate
    AFTER INSERT
    ON feed_items
BEGIN
    INSERT INTO feed_items_fts5(title, id) VALUES (NEW.title, NEW.id);
end;

INSERT INTO feed_items_fts5(title, id)
SELECT title, id from feed_items;

VACUUM;
