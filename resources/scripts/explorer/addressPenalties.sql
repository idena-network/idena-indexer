select a.address, p.penalty, p.block_height, b.hash, b.timestamp, b.epoch
from penalties p
         join blocks b on b.height = p.block_height
         join addresses a on a.id = p.address_id
where lower(a.address) = lower($1)
order by p.block_height desc
limit $3
    offset $2