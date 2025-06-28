-- Main store of NZB feed/search items
CREATE TABLE feed_items
(
    id         TEXT PRIMARY KEY, -- indexer_name:GUID
    indexer_name TEXT NOT NULL,
    title      TEXT    NOT NULL,
    artist     TEXT,
    guid       TEXT,             -- original feed GUID (may not be unique across indexers)
    link       TEXT,             -- optional HTML page
    nzb_url    TEXT    NOT NULL,
    pub_date   DATETIME,
    size       INTEGER,          -- in bytes
    category   TEXT,
    source     TEXT    NOT NULL, -- 'rss' or 'search'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Track past search queries to avoid repetition (per indexer)
CREATE TABLE search_cache
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    indexer_name  TEXT NOT NULL,
    query         TEXT     NOT NULL,
    categories    TEXT     NOT NULL,
    first_tried   DATETIME NOT NULL,
    last_tried    DATETIME NOT NULL,
    status        TEXT     NOT NULL, -- 'hit', 'miss', 'error'
    error_message TEXT,
    UNIQUE (indexer_name, query)
);

-- Optionally track local cache of NZB downloads
CREATE TABLE nzb_cache
(
    feed_item_id TEXT PRIMARY KEY REFERENCES feed_items (id),
    filename     TEXT NOT NULL, -- local path or filename
    saved_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);
