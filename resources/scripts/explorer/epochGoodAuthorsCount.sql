select count(*)
from good_authors ga
         join epoch_identities ei on ei.id = ga.epoch_identity_id
where ei.epoch = $1