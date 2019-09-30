select count(*)
from mining_rewards mr
         join addresses a on a.id = mr.address_id
where lower(a.address) = lower($1)