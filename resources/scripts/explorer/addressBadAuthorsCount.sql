select count(*)
from bad_authors ba
         join address_states s on s.id = ba.ei_address_state_id
         join addresses a on a.id = s.address_id
where lower(a.address) = lower($1)