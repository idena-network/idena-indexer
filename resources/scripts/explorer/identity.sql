select dis.name state
from addresses a
         join address_states s on s.address_id = a.id and s.is_actual
         join dic_identity_states dis on dis.id = s.state
where LOWER(a.address) = LOWER($1)