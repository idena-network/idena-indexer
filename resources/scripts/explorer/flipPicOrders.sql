select fpo.answer_index, fpo.flip_pic_index
from flip_pic_orders fpo
where fpo.fd_flip_tx_id = (select tx_id from flips where lower(cid) = lower($1))
order by fpo.pos_index