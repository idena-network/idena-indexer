select ''                         address,
       0                          epoch,
       ba.reason = 2              reported,
       dicr.name                  reason,
       coalesce(prevdis.name, '') prev_state,
       dis.name                   state
from bad_authors ba
         join dic_bad_author_reasons dicr on dicr.id = ba.reason
         join epoch_identities ei on ei.address_state_id = ba.ei_address_state_id and ei.epoch = $1
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($2)
         left join address_states prevs on prevs.id = s.prev_id
         join dic_identity_states dis on dis.id = s.state
         left join dic_identity_states prevdis on prevdis.id = prevs.state