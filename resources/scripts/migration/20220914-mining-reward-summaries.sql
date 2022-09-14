CREATE TABLE IF NOT EXISTS mining_reward_summaries
(
    address_id bigint          NOT NULL,
    epoch      smallint        NOT NULL,
    amount     numeric(30, 18) NOT NULL,
    burnt      numeric(30, 18) NOT NULL
);

INSERT INTO mining_reward_summaries (address_id, epoch, amount, burnt)
SELECT bu.address_id,
       b.epoch,
       sum(bu.balance_new - bu.balance_old + bu.stake_new - bu.stake_old +
           (coalesce(bu.penalty_old, 0) - coalesce(bu.penalty_new, 0)) + coalesce(bu.penalty_payment, 0)) amount,
       sum((coalesce(bu.penalty_old, 0) - coalesce(bu.penalty_new, 0)) + coalesce(bu.penalty_payment, 0)) burnt
FROM balance_updates bu
         JOIN blocks b ON b.height = bu.block_height
WHERE bu.reason IN (2, 3)
GROUP BY bu.address_id, b.epoch;