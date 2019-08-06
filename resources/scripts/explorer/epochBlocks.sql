select b.height,
       b.Hash,
       b.timestamp,
       (select count(*) from transactions where block_id = b.id) TX_COUNT,
       a.address                                                 proposer
from blocks b
         join epochs e on e.id = b.epoch_id
         join proposers p on p.block_id = b.id
         join addresses a on a.id = p.address_id
where e.epoch = $1
limit $3
    offset $2