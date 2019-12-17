select dis.name            state,
       b.epoch,
       s.block_height,
       b.hash,
       coalesce(t.hash, ''),
       b.timestamp,
       (ei.id is not null) is_validation
from address_states s
         join blocks b on b.height = s.block_height
         join addresses a on a.id = s.address_id
         left join epoch_identities ei on ei.address_state_id = s.id
         left join transactions t on t.id = s.tx_id
         join dic_identity_states dis on dis.id = s.state
where lower(a.address) = lower($1)
order by s.block_height desc
limit $3
offset
$2