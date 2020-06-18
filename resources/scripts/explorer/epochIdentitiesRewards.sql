select a.address,
       vr.balance,
       vr.stake,
       dert.name                  "type",
       coalesce(prevdis.name, '') prev_state,
       dis.name                   state,
       coalesce(ra.age, 0)
from validation_rewards vr
         join address_states s on s.id = vr.ei_address_state_id
         join addresses a on a.id = s.address_id
         join (select ei.address_state_id, totals.total_reward
               from epoch_identities ei
                        join (select ei_address_state_id, sum(balance + stake) total_reward
                              from validation_rewards
                              group by ei_address_state_id) totals
                             on totals.ei_address_state_id = ei.address_state_id
               where ei.epoch = $1
               order by totals.total_reward desc, ei.address_state_id
               limit $2 offset $3) filtered on filtered.address_state_id = vr.ei_address_state_id
         left join address_states prevs on prevs.id = s.prev_id
         left join reward_ages ra on ra.ei_address_state_id = vr.ei_address_state_id
         join dic_identity_states dis on dis.id = s.state
         left join dic_identity_states prevdis on prevdis.id = prevs.state
         join dic_epoch_reward_types dert on dert.id = vr.type
order by filtered.total_reward desc, filtered.address_state_id