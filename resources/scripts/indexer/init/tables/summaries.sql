CREATE TABLE IF NOT EXISTS epoch_summaries
(
    epoch                bigint          NOT NULL,
    validated_count      integer         NOT NULL,
    block_count          bigint          NOT NULL,
    empty_block_count    bigint          NOT NULL,
    tx_count             bigint          NOT NULL,
    invite_count         bigint          NOT NULL,
    flip_count           integer         NOT NULL,
    burnt                numeric(30, 18) NOT NULL,
    minted               numeric(30, 18) NOT NULL,
    total_balance        numeric(30, 18) NOT NULL,
    total_stake          numeric(30, 18) NOT NULL,
    block_height         bigint          NOT NULL,
    min_score_for_invite real            NOT NULL,
    CONSTRAINT epoch_summaries_pkey PRIMARY KEY (epoch),
    CONSTRAINT epoch_summaries_block_height_fkey FOREIGN KEY (block_height)
        REFERENCES blocks (height) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT epoch_summaries_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS flip_summaries
(
    flip_tx_id        bigint   NOT NULL,
    wrong_words_votes smallint NOT NULL,
    short_answers     smallint NOT NULL,
    long_answers      smallint NOT NULL,
    encrypted         boolean  NOT NULL,
    CONSTRAINT flip_summaries_pkey PRIMARY KEY (flip_tx_id),
    CONSTRAINT flip_summaries_flip_tx_id_fkey FOREIGN KEY (flip_tx_id)
        REFERENCES flips (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS coins_summary
(
    total_burnt  numeric(30, 18),
    total_minted numeric(30, 18)
);