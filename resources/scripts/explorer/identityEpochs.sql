select a.address,
       eis.epoch,
       eis.state,
       coalesce(prevs.state, '') prev_state,
       coalesce(ei.approved, false),
       coalesce(ei.missed, false),
       coalesce(ei.short_point, 0),
       coalesce(ei.short_flips, 0),
       coalesce(ei.total_short_point, 0),
       coalesce(ei.total_short_flips, 0),
       coalesce(ei.long_point, 0),
       coalesce(ei.long_flips, 0),
       coalesce(ei.required_flips, 0) required_flips,
       coalesce(ei.made_flips, 0) made_flips
from epoch_identity_states eis
         join addresses a on a.id = eis.address_id
         left join address_states prevs on prevs.id = eis.prev_id
         left join epoch_identities ei on ei.epoch = eis.epoch and ei.address_state_id = eis.address_state_id
where lower(a.address) = lower($1)
order by eis.epoch desc
limit $3
    offset $2