CREATE TABLE IF NOT EXISTS miners_history
(
    block_timestamp   bigint NOT NULL,
    online_validators bigint NOT NULL,
    online_miners     bigint NOT NULL,
    CONSTRAINT miners_history_pkey PRIMARY KEY (block_timestamp)
);