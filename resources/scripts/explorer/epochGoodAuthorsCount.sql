select count(*)
from good_authors ga
         join epoch_identities ei on ei.address_state_id = ga.ei_address_state_id
where ei.epoch = $1