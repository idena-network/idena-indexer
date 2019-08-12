select f.cid
from flips_to_solve fts
         join flips f on f.id = fts.flip_id
         join epoch_identities ei on ei.id = fts.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
where ei.epoch = $1
  and lower(a.address) = lower($2)
  and fts.is_short = $3