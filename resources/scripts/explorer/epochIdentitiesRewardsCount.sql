select count(*)
from (select distinct vr.ei_address_state_id
      from validation_rewards vr
               join epoch_identities ei on ei.address_state_id = vr.ei_address_state_id and ei.epoch = $1) vr_eid