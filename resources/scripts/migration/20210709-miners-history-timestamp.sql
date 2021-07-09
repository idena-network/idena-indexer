ALTER TABLE miners_history
    RENAME TO miners_history_old;

ALTER TABLE miners_history_old
    RENAME CONSTRAINT miners_history_pkey TO miners_history_old_pkey;

CREATE TABLE IF NOT EXISTS miners_history
(
    block_timestamp   bigint NOT NULL,
    online_validators bigint NOT NULL,
    online_miners     bigint NOT NULL,
    CONSTRAINT miners_history_pkey PRIMARY KEY (block_timestamp)
);

INSERT INTO miners_history
SELECT b.timestamp, mh.online_validators, mh.online_miners
FROM miners_history_old mh
         JOIN blocks b ON b.height = mh.block_height
ORDER BY mh.block_height;

DROP TABLE miners_history_old;