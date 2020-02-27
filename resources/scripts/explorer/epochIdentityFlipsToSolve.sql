select f.cid
from flips_to_solve fts
         join flips f on f.tx_id = fts.flip_tx_id
         join epoch_identities ei on ei.address_state_id = fts.ei_address_state_id and ei.epoch = $1
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($2)
where fts.is_short = $3