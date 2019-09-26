select count(*)
from (select distinct vr.epoch_identity_id
      from validation_rewards vr
               join epoch_identities ei on ei.id = vr.epoch_identity_id
      where ei.epoch = $1) vr_eid