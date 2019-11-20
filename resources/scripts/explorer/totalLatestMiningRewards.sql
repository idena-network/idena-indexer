select a.address,
       tlr.balance,
       tlr.stake,
       proposer,
       final_committee
from (select lr.address_id,
             sum(lr.balance)      balance,
             sum(lr.stake)        stake,
             sum(proposer)        proposer,
             sum(final_committee) final_committee
      from (select mr.address_id,
                   mr.balance,
                   mr.stake,
                   (case when mr."type" = 'Proposer' then 1 else 0 end)       proposer,
                   (case when mr."type" = 'FinalCommittee' then 1 else 0 end) final_committee
            from mining_rewards mr
                     join blocks b on b.height = mr.block_height
            where b."timestamp" > $1) lr
      group by lr.address_id) tlr
         join addresses a on a.id = tlr.address_id
order by tlr.balance + tlr.stake desc
limit $3
offset
$2