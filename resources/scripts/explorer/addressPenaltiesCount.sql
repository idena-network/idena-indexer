select count(*)
from penalties p
         join addresses a on a.id = p.address_id
where lower(a.address) = lower($1)