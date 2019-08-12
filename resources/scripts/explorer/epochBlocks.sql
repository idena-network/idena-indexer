select b.height,
       b.Hash,
       b.timestamp,
       (select count(*) from transactions where block_height = b.height) TX_COUNT,
       coalesce(a.address, '')                                           proposer
from blocks b
         left join proposers p on p.block_height = b.height
         left join addresses a on a.id = p.address_id
where b.epoch = $1
order by b.height desc
limit $3
    offset $2