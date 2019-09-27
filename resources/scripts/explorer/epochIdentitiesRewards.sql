select a.address,
       vr.balance,
       vr.stake,
       vr.type,
       coalesce(prev_states.state) prev_state,
       s.state
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
         left join address_states prev_states on prev_states.id = s.prev_id
order by filtered.total_reward desc, a.address