select i.address, a.answer, a.is_short
from answers a
         join identities i on i.id = a.identity_id
where a.flip_id = $1