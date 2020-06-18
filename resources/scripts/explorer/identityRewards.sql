select a.address,
       ei.epoch,
       0         block_height,
       vr.balance,
       vr.stake,
       dert.name "type"
from validation_rewards vr
         join epoch_identities ei on ei.address_state_id = vr.ei_address_state_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($1)
         join dic_epoch_reward_types dert on dert.id = vr.type
order by vr.ei_address_state_id desc, vr.type desc
limit $2 offset $3