-- old migration
DROP PROCEDURE migrate_balance_updates;

-- total_rewards
ALTER TABLE total_rewards
    RENAME TO total_rewards_old;
ALTER TABLE total_rewards_old
    RENAME CONSTRAINT total_rewards_pkey TO total_rewards_pkey_old;
ALTER TABLE total_rewards_old
    RENAME CONSTRAINT total_rewards_block_height_fkey TO total_rewards_block_height_fkey_old;

CREATE TABLE IF NOT EXISTS total_rewards
(
    epoch             bigint          NOT NULL,
    total             numeric(30, 18) NOT NULL,
    validation        numeric(30, 18) NOT NULL,
    flips             numeric(30, 18) NOT NULL,
    invitations       numeric(30, 18) NOT NULL,
    foundation        numeric(30, 18) NOT NULL,
    zero_wallet       numeric(30, 18) NOT NULL,
    validation_share  numeric(30, 18) NOT NULL,
    flips_share       numeric(30, 18) NOT NULL,
    invitations_share numeric(30, 18) NOT NULL,
    CONSTRAINT total_rewards_pkey PRIMARY KEY (epoch),
    CONSTRAINT total_rewards_epoch_fkey FOREIGN KEY (epoch)
        REFERENCES epochs (epoch) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

INSERT INTO total_rewards
SELECT b.epoch,
       tro.total,
       tro.validation,
       tro.flips,
       tro.invitations,
       tro.foundation,
       tro.zero_wallet,
       tro.validation_share,
       tro.flips_share,
       tro.invitations_share
FROM total_rewards_old tro
         JOIN blocks b on b.height = tro.block_height
ORDER BY b.epoch;

DROP TABLE total_rewards_old;

DROP PROCEDURE save_total_reward;

-- flip_lottery_block_height
ALTER TABLE epoch_summaries
    ADD COLUMN flip_lottery_block_height bigint;

DROP PROCEDURE update_epoch_summary;
DROP PROCEDURE save_block;

UPDATE epoch_summaries
SET flip_lottery_block_height = t.height
FROM (SELECT b.epoch, b.height
      FROM blocks b
               JOIN block_flags bf ON bf.block_height = b.height AND bf.flag = 'FlipLotteryStarted') t
WHERE epoch_summaries.epoch = t.epoch;

-- epoch min and max tx id
ALTER TABLE epoch_summaries
    ADD COLUMN min_tx_id bigint;
ALTER TABLE epoch_summaries
    ADD COLUMN max_tx_id bigint;

UPDATE epoch_summaries
SET min_tx_id = t.min_tx_id,
    max_tx_id = t.max_tx_id
FROM (SELECT b.epoch, min(t.id) min_tx_id, max(t.id) max_tx_id
      FROM transactions t
               JOIN blocks b on b.height = t.block_height
      GROUP BY b.epoch) t
WHERE epoch_summaries.epoch = t.epoch;

-- epoch flip statuses
CREATE TABLE IF NOT EXISTS epoch_flip_statuses
(
    epoch       bigint   NOT NULL,
    flip_status smallint NOT NULL,
    count       integer  NOT NULL,
    CONSTRAINT epoch_flip_statuses_pkey PRIMARY KEY (epoch, flip_status)
);

INSERT INTO epoch_flip_statuses
SELECT b.epoch, dfs.id, count(*)
FROM flips f
         JOIN transactions t ON t.id = f.tx_id
         JOIN blocks b ON b.height = t.block_height
         JOIN dic_flip_statuses dfs ON dfs.id = f.status
WHERE f.delete_tx_id is null
GROUP BY b.epoch, dfs.id
ORDER BY b.epoch, dfs.id;

DROP PROCEDURE save_epoch_result;
DROP PROCEDURE save_epoch_rewards_bounds;
DROP FUNCTION save_addrs_and_txs;

-- epoch reported flips
ALTER TABLE epoch_summaries
    ADD COLUMN reported_flips integer;

UPDATE epoch_summaries
SET reported_flips = t.cnt
FROM (select b.epoch, f.grade, count(*) cnt
      from flips f
               join transactions t on t.id = f.tx_id
               join blocks b on b.height = t.block_height
      where f.delete_tx_id is null
      group by b.epoch, f.grade) t
WHERE epoch_summaries.epoch = t.epoch
  AND t.grade = 1;

DROP PROCEDURE generate_epoch_summaries;