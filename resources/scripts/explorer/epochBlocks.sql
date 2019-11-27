select b.height,
       b.Hash,
       b.timestamp,
       (select count(*) from transactions where block_height = b.height) TX_COUNT,
       coalesce(a.address, '')                                           proposer,
       b.is_empty,
       b.size,
       b.vrf_proposer_threshold,
       c.burnt,
       c.minted,
       c.total_balance,
       c.total_stake
from blocks b
         left join block_proposers p on p.block_height = b.height
         left join addresses a on a.id = p.address_id
         join coins c on c.block_height = b.height
where b.epoch = $1
order by b.height desc
limit $3
offset
$2