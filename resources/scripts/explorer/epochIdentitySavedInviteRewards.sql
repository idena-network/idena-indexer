select dicrt.name, sir.count
from epoch_identities ei
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id and lower(a.address) = lower($2)
         join saved_invite_rewards sir on sir.ei_address_state_id = ei.address_state_id
         join dic_epoch_reward_types dicrt on dicrt.id = sir.reward_type
where ei.epoch = $1