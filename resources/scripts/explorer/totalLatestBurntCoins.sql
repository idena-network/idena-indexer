select a.address,
       tlbc.amount
from (select lbc.address_id,
             sum(lbc.amount) amount
      from (select bc.address_id,
                   bc.amount
            from burnt_coins bc
                     join blocks b on b.height = bc.block_height
            where b."timestamp" > $1) lbc
      group by lbc.address_id) tlbc
         join addresses a on a.id = tlbc.address_id
order by tlbc.amount desc
limit $3
offset
$2
