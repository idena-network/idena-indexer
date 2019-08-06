select a.address,
       eis.epoch,
       eis.state,
       coalesce((
                    select as2.state
                    from address_states as2
                             join blocks b on b.id = as2.block_Id
                    where as2.address_id = eis.address_id
                      and b.height < (select height from blocks where id = eis.block_id)
                    order by b.height desc
                    limit 1
                ), '') prev_state,
       coalesce(ei.approved, false),
       coalesce(ei.missed, false),
       coalesce(ei.short_point, 0),
       coalesce(ei.short_flips, 0),
       coalesce(ei.total_short_point, 0),
       coalesce(ei.total_short_flips, 0),
       coalesce(ei.long_point, 0),
       coalesce(ei.long_flips, 0)
from epoch_identity_states eis
         join addresses a on a.id = eis.address_id
         left join epoch_identities ei on ei.epoch_id = eis.epoch_id and ei.address_state_id = eis.address_state_id
where eis.epoch = $1
limit $3
    offset $2