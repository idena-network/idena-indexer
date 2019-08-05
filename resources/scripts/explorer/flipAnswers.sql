select ad.address, a.answer, coalesce(f.answer)
from answers a
         join epoch_identities ei on ei.id = a.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses ad on ad.id = s.address_id
         join flips f on f.id = a.flip_id
where lower(f.cid) = lower($1)
  and a.is_short = $2
limit $4
    offset $3