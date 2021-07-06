CREATE TABLE IF NOT EXISTS peers_history
(
    timestamp bigint NOT NULL,
    count     bigint NOT NULL,
    CONSTRAINT peers_history_pkey PRIMARY KEY (timestamp)
);