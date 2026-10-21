CREATE TABLE metrics (
    "id"        TEXT NOT NULL,
    "type"      TEXT NOT NULL,
    "delta"     BIGINT,
    "value"     DOUBLE PRECISION,
    "hash"      TEXT,
    PRIMARY KEY ("id", "type")
);