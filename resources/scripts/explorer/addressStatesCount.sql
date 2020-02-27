select count(*)
from address_states s
         join addresses a on a.id = s.address_id and lower(a.address) = lower($1)