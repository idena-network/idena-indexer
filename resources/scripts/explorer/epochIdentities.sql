SELECT eis.address_state_id,
       a.address,
       eis.epoch,
       dis.name                                                       state,
       coalesce(prevdis.name, '')                                     prev_state,
       coalesce(ei.approved, false),
       coalesce(ei.missed, false),
       coalesce(ei.short_point, 0),
       coalesce(ei.short_flips, 0),
       coalesce(least(ei.total_short_point, ei.total_short_flips), 0) total_short_point,
       coalesce(ei.total_short_flips, 0),
       coalesce(ei.long_point, 0),
       coalesce(ei.long_flips, 0),
       coalesce(ei.required_flips, 0)                                 required_flips,
       coalesce(ei.made_flips, 0)                                     made_flips,
       coalesce(ei.available_flips, 0)                                available_flips,
       coalesce((SELECT sum(vr.balance + vr.stake)
                 FROM validation_rewards vr
                 WHERE vr.ei_address_state_id = eis.address_state_id
                   AND ei.epoch = eis.epoch), 0)                      total_validation_reward,
       coalesce(ei.birth_epoch, 0)                                    birth_epoch,
       coalesce(ei.short_answers, 0)                                  short_answers,
       coalesce(ei.long_answers, 0)                                   long_answers
FROM epoch_identity_states eis
         JOIN addresses a ON a.id = eis.address_id
         LEFT JOIN address_states prevs ON prevs.id = eis.prev_id
         LEFT JOIN epoch_identities ei ON ei.epoch = eis.epoch AND ei.address_state_id = eis.address_state_id
         JOIN dic_identity_states dis ON dis.id = eis.state
         LEFT JOIN dic_identity_states prevdis ON prevdis.id = prevs.state
WHERE ($5::bigint IS NULL OR eis.address_state_id >= $5)
  AND eis.epoch = $1
  AND ($2::smallint[] IS NULL OR prevs.state = ANY ($2::smallint[]))
  AND ($3::smallint[] IS NULL OR eis.state = ANY ($3::smallint[]))
ORDER BY eis.address_state_id
LIMIT $4