select eis.state,
       coalesce(preva.state, '') prev_state,
       coalesce(ei.short_point, 0),
       coalesce(ei.short_flips, 0),
       coalesce(ei.total_short_point, 0),
       coalesce(ei.total_short_flips, 0),
       coalesce(ei.long_point, 0),
       coalesce(ei.long_flips, 0),
       coalesce(ei.approved, false),
       coalesce(ei.missed, false)
from epoch_identity_states eis
         left join epoch_identities ei on ei.epoch = eis.epoch and ei.address_state_id = eis.address_state_id
         left join address_states preva on preva.id = eis.prev_id
         join addresses a on a.id = eis.address_id
where eis.epoch = $1
  and LOWER(a.address) = LOWER($2)