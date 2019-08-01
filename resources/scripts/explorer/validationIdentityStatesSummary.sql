select s.state, count(*)
from address_states s
         join epoch_identities ei on ei.address_state_id = s.id
         join epochs e on e.id = ei.epoch_id
where e.epoch = $1
group by s.state