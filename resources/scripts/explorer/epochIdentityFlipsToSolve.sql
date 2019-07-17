select f.cid, fts.is_short
from flips_to_solve fts
         join flips f on f.id = fts.flip_id
where fts.epoch_identity_id = $1