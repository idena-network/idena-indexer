select count(*)
from validation_rewards vr
         join address_states s on s.id = vr.ei_address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($1)