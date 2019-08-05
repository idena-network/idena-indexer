select count(*) state_count
from address_states s
         join addresses a on a.id = s.address_id
where lower(a.address) = lower($1)