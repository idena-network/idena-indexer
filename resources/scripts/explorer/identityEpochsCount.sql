select count(*) epoch_count
from epoch_identity_states eis
         join addresses a on a.id = eis.address_id
where lower(a.address) = lower($1)