select dis.name state,
       count(*) cnt
from epoch_identities ei
         join address_states s on s.id = ei.address_state_id
         join dic_identity_states dis on dis.id = s.state
where ei.epoch = $1
group by dis.name;