select a.address,
       b.epoch,
       mr.block_height,
       mr.balance,
       mr.stake,
       mr.type
from mining_rewards mr
         join addresses a on a.id = mr.address_id
         join blocks b on b.height = mr.block_height
where lower(a.address) = lower($1)
order by mr.block_height desc
limit $3
offset
$2