select count(*) identity_count
from epoch_identity_states eis
where eis.epoch = $1