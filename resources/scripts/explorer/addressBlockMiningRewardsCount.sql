select count(*)
from (select distinct mr.block_height
      from mining_rewards mr
               join addresses a on a.id = mr.address_id
      where lower(a.address) = lower($1)) ba