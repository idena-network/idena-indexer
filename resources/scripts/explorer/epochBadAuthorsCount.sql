select count(*)
from bad_authors ba
         join epoch_identities ei on ei.address_state_id = ba.ei_address_state_id
where ei.epoch = $1