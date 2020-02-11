select b.height,
       b.Hash,
       b.timestamp,
       (select count(*) from transactions where block_height = b.height) TX_COUNT,
       coalesce(a.address, '')                                           proposer,
       coalesce(vs.vrf_score, 0)                                         proposer_vrf_score,
       b.is_empty,
       b.body_size,
       b.full_size,
       b.vrf_proposer_threshold,
       b.fee_rate,
       c.burnt,
       c.minted,
       c.total_balance,
       c.total_stake,
       bf.flags
from (select *
      from blocks
      where epoch = $1
      order by height desc
      limit $3
      offset
      $2) b
         left join block_proposers p on p.block_height = b.height
         left join block_proposer_vrf_scores vs on vs.block_height = b.height
         left join addresses a on a.id = p.address_id
         join coins c on c.block_height = b.height
         left join (select block_height, array_agg("flag") flags
                    from block_flags
                    group by block_height) bf on bf.block_height = b.height
order by b.height desc