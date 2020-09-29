select least(ei.total_short_point, ei.total_short_flips) total_short_point,
       ei.total_short_flips
from epoch_identities ei
         join address_states s on s.id = ei.address_state_id and
                                  s.address_id = (select id from addresses where lower(address) = lower($1))
order by ei.address_state_id desc
limit 1