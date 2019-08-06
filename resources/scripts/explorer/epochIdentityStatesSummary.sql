select s.state, count(*)
from address_states s
         join epoch_identity_states eis on eis.epoch = $1 and eis.address_state_id = s.id
group by s.state;