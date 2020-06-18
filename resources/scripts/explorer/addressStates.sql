select s.id,
       dis.name                          state,
       b.epoch,
       s.block_height,
       b.hash,
       coalesce(t.hash, ''),
       b.timestamp,
       (ei.address_state_id is not null) is_validation
from address_states s
         join addresses a on a.id = s.address_id and lower(a.address) = lower($1)
         join blocks b on b.height = s.block_height
         left join epoch_identities ei on ei.address_state_id = s.id
         left join transactions t on t.id = s.tx_id
         join dic_identity_states dis on dis.id = s.state
WHERE $3::bigint IS NULL
   OR s.id <= $3
order by s.id desc
limit $2