select ba.ei_address_state_id     id,
       ''                         address,
       ei.epoch                   epoch,
       ba.reason = 2              reported,
       dicr.name                  reason,
       coalesce(prevdis.name, '') prev_state,
       dis.name                   state
from bad_authors ba
         join address_states s on s.id = ba.ei_address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($1)
         join epoch_identities ei on ei.address_state_id = ba.ei_address_state_id
         join dic_bad_author_reasons dicr on dicr.id = ba.reason
         join dic_identity_states dis on dis.id = s.state
         left join address_states prevs on prevs.id = s.prev_id
         left join dic_identity_states prevdis on prevdis.id = prevs.state
WHERE $3::bigint IS NULL
   OR ba.ei_address_state_id <= $3
order by ba.ei_address_state_id desc
limit $2