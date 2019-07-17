select i.id,
       i.address,
       (select state from epoch_identities where identity_id = i.id order by epoch_id desc limit 1) state
from identities i
where LOWER(i.address) = LOWER($1)
