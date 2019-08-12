select b.height,
       b.hash,
       b.timestamp,
       (select count(*) from transactions where block_height = b.height) TX_COUNT,
       coalesce(pa.address, '')                                          proposer
from blocks b
         left join proposers p on p.block_height = b.height
         left join addresses pa on pa.id = p.address_id
where lower(b.hash) = lower($1)