select count(*)
from validation_rewards vr
         join epoch_identities ei on ei.id = vr.epoch_identity_id
where ei.epoch = $1