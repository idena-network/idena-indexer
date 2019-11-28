select count(*) identity_count
from epoch_identity_states eis
where eis.epoch = $1
  and ($2::text[] is null or eis.state = any ($2::text[]))