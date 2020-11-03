select b.epoch,
       b.height,
       b.hash,
       b.timestamp,
       (select count(*) from transactions where block_height = b.height)         TX_COUNT,
       b.validators_count,
       coalesce(pa.address, '')                                                  proposer,
       coalesce(vs.vrf_score, 0)                                                 proposer_vrf_score,
       b.is_empty,
       b.body_size,
       b.full_size,
       b.vrf_proposer_threshold,
       b.fee_rate,
       (select array_agg("flag") from block_flags where block_height = b.height) flags,
       b.upgrade
from blocks b
         left join block_proposers p on p.block_height = b.height
         left join block_proposer_vrf_scores vs on vs.block_height = b.height
         left join addresses pa on pa.id = p.address_id
where lower(b.hash) = lower($1)