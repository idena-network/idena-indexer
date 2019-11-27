select b.height,
       b.hash,
       b.timestamp,
       (select count(*) from transactions where block_height = b.height) TX_COUNT,
       b.validators_count,
       coalesce(pa.address, '')                                          proposer,
       b.is_empty,
       b.size,
       b.vrf_proposer_threshold
from blocks b
         left join block_proposers p on p.block_height = b.height
         left join addresses pa on pa.id = p.address_id
where b.height = $1