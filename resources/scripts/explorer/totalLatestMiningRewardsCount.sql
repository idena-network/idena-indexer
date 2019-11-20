select count(*)
from (select distinct mr.address_id
      from mining_rewards mr
               join blocks b on b.height = mr.block_height
      where b."timestamp" > $1) lr