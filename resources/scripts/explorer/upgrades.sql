SELECT b.height,
       b.Hash,
       b.timestamp,
       (SELECT count(*) FROM transactions WHERE block_height = b.height)         TX_COUNT,
       coalesce(a.address, '')                                                   proposer,
       coalesce(vs.vrf_score, 0)                                                 proposer_vrf_score,
       b.is_empty,
       b.body_size,
       b.full_size,
       b.vrf_proposer_threshold,
       b.fee_rate,
       c.burnt,
       c.minted,
       c.total_balance,
       c.total_stake,
       (SELECT array_agg("flag") FROM block_flags WHERE block_height = b.height) flags,
       b.upgrade
FROM blocks b
         LEFT JOIN block_proposers p ON p.block_height = b.height
         LEFT JOIN block_proposer_vrf_scores vs ON vs.block_height = b.height
         LEFT JOIN addresses a ON a.id = p.address_id
         JOIN coins c ON c.block_height = b.height
WHERE coalesce(upgrade, 0) > 0
  AND ($2::bigint IS NULL OR height <= $2::bigint)
ORDER BY b.height DESC
LIMIT $1