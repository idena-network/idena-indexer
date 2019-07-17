select i.address, a.answer, a.is_short
from answers a
         join epoch_identities ei on ei.id = a.epoch_identity_id
         join identities i on i.id = ei.identity_id
where a.flip_id = $1