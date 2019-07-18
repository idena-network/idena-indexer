select ei.id
from epoch_identities ei
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
         join epochs e on e.id = ei.epoch_id
where e.epoch = $1
  and LOWER(a.address) = LOWER($2)