select count(*)
from bad_authors ba
         join epoch_identities ei on ei.id = ba.epoch_identity_id
where ei.epoch = $1