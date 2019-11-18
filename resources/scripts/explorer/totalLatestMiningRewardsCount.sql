select count(*)
from (select distinct mr.address_id
      from mining_rewards mr
      where mr.block_height > (select max(height) from blocks) - $1) lr