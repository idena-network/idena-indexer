select a.address,
       vr.balance,
       vr.stake,
       dert.name                  "type",
       coalesce(prevdis.name, '') prev_state,
       dis.name                   state,
       coalesce(ra.age, 0)
from validation_rewards vr
         join epoch_identities ei on ei.id = vr.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
         join (select distinct epoch_identity_id, total_reward
               from (select vr.epoch_identity_id, totals.total_reward
                     from validation_rewards vr
                              join epoch_identities ei on ei.id = vr.epoch_identity_id
                              join (select epoch_identity_id, sum(balance + stake) total_reward
                                    from validation_rewards
                                    group by epoch_identity_id) totals
                                   on totals.epoch_identity_id = vr.epoch_identity_id
                     where ei.epoch = $1
                     order by totals.total_reward desc, vr.epoch_identity_id
                    ) sorted
               order by total_reward desc
               limit $3
               offset
               $2) filtered on filtered.epoch_identity_id = vr.epoch_identity_id
         left join address_states prevs on prevs.id = s.prev_id
         left join reward_ages ra on ra.epoch_identity_id = vr.epoch_identity_id
         join dic_identity_states dis on dis.id = s.state
         left join dic_identity_states prevdis on prevdis.id = prevs.state
         join dic_epoch_reward_types dert on dert.id = vr.type
order by filtered.total_reward desc, a.address