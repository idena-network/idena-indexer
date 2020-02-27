with cur_epoch as (
    select burnt, minted, total_balance, total_stake
    from (SELECT COALESCE(sum(c.minted), 0), COALESCE(sum(c.burnt), 0)
          FROM coins c
                   JOIN blocks b ON b.height = c.block_height
          WHERE b.epoch = $1) t1 (minted, burnt),
         (SELECT c.total_balance, c.total_stake
          FROM coins c
                   JOIN blocks b ON b.height = c.block_height
          WHERE b.epoch = $1
          ORDER BY c.block_height DESC
          LIMIT 1) t2
)
select coalesce(sum(burnt), (select burnt from cur_epoch))                 burnt,
       coalesce(sum(minted), (select minted from cur_epoch))               minted,
       coalesce(sum(total_balance), (select total_balance from cur_epoch)) total_balance,
       coalesce(sum(total_stake), (select total_stake from cur_epoch))     total_stake
from epoch_summaries
where epoch = $1