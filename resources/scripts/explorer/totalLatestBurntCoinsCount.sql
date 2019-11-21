select count(*)
from (select distinct bc.address_id
      from burnt_coins bc
               join blocks b on b.height = bc.block_height
      where b."timestamp" > $1) lbc