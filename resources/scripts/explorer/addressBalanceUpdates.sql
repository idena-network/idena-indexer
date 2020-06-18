select bu.id,
       bu.balance_old,
       bu.stake_old,
       coalesce(bu.penalty_old, 0)            penalty_old,
       bu.balance_new,
       bu.stake_new,
       coalesce(bu.penalty_new, 0)            penalty_new,
       dicr.name                              reason,
       b.height                               block_height,
       b.hash                                 block_hash,
       b.timestamp                            block_timestamp,
       coalesce(t.hash, '')                   tx_hash,
       coalesce(lb.height, 0)                 last_block_height,
       coalesce(lb.hash, '')                  last_block_hash,
       coalesce(lb.timestamp, 0)              last_block_timestamp,
       coalesce(bu.committee_reward_share, 0) committee_reward_share,
       coalesce(bu.blocks_count, 0)           blocks_count
from balance_updates bu
         left join transactions t on t.id = bu.tx_id
         join dic_balance_update_reasons dicr on dicr.id = bu.reason
         join blocks b on b.height = bu.block_height
         left join blocks lb on lb.height = bu.last_block_height
where ($3::bigint IS NULL OR bu.id <= $3)
  AND bu.address_id = (select id from addresses where lower(address) = lower($1))
order by bu.id desc
limit $2