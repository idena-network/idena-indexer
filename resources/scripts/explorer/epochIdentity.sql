select ei.id
from epoch_identities ei
         join identities i on i.id = ei.identity_id
         join epochs e on e.id = ei.epoch_id
where e.epoch = $1
  and LOWER(i.address) = LOWER($2)