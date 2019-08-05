select count(*) identity_count
from epoch_identities ei
         join epochs e on e.id = ei.epoch_id
where e.epoch = $1