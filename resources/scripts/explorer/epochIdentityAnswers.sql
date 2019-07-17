select f.cid, a.answer, a.is_short
from answers a
         join flips f on f.id = a.flip_id
where a.epoch_identity_id = $1