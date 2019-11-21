select coalesce(sum(bc.amount), 0) amount
from burnt_coins bc
         join addresses a on a.id = bc.address_id
         join blocks b on b.height = bc.block_height
where b."timestamp" > $1
  and lower(a.address) = lower($2);