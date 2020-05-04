select count(*) epoch_count
from epoch_identities ei
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($1);