select a.address,
       p.penalty,
       coalesce(pp.penalty, 0) paid,
       p.block_height,
       b.hash,
       b.timestamp,
       b.epoch
from penalties p
         join blocks b on b.height = p.block_height
         join addresses a on a.id = p.address_id
         left join paid_penalties pp on pp.penalty_id = p.id
where lower(a.address) = lower($1)
order by p.id desc
limit $3 offset $2