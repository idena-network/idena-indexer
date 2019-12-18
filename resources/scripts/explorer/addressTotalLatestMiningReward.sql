select coalesce(sum(mr.balance), 0)                              balance,
       coalesce(sum(mr.stake), 0)                                stake,
       coalesce(sum(case when mr.proposer then 1 else 0 end), 0) proposer,
       coalesce(sum(case when mr.proposer then 0 else 1 end), 0) final_committee
from mining_rewards mr
         join addresses a on a.id = mr.address_id
         join blocks b on b.height = mr.block_height
where b."timestamp" > $1
  and lower(a.address) = lower($2);