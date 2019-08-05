select s.state
from addresses a
         join address_states s on s.address_id = a.id and s.is_actual
where LOWER(a.address) = LOWER($1)