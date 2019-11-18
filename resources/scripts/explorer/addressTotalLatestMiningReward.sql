select coalesce(sum(mr.balance), 0)                                  balance,
       coalesce(sum(mr.stake), 0)                                    stake,
       sum(case when mr."type" = 'Proposer' then 1 else 0 end)       proposer,
       sum(case when mr."type" = 'FinalCommittee' then 1 else 0 end) final_committee
from mining_rewards mr
         join addresses a on a.id = mr.address_id
where mr.block_height > (select max(height) from blocks) - $1
  and lower(a.address) = lower($2);