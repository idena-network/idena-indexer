select count(*) identity_count
from epoch_identity_states eis
where eis.epoch = $1
  and ($2::smallint[] is null or eis.state = any ($2::smallint[]))