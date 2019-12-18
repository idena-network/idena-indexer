select mr.block_height,
       b.epoch,
       mr.balance,
       mr.stake,
       (case when mr.proposer then 'Proposer' else 'FinalCommittee' end) "type"
from mining_rewards mr
         join (select distinct mr.block_height, mr.address_id
               from mining_rewards mr
                        join addresses a on a.id = mr.address_id
               where lower(a.address) = lower($1)
               order by mr.block_height desc
               limit $3
               offset
               $2) ba on ba.block_height = mr.block_height and ba.address_id = mr.address_id
         join blocks b on b.height = mr.block_height
order by mr.block_height desc