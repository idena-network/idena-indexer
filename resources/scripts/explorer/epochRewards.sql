select a.address,
       ei.epoch,
       0 block_height,
       vr.balance,
       vr.stake,
       vr.type
from validation_rewards vr
         join epoch_identities ei on ei.id = vr.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
where ei.epoch = $1
limit $3
offset
$2