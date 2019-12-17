select dis.name state,
       count(*) cnt
from address_states s
         join epoch_identity_states eis on eis.epoch = $1 and eis.address_state_id = s.id
         join dic_identity_states dis on dis.id = s.state
group by dis.name;