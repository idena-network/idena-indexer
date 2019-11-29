select fpo.answer_index, fpo.flip_pic_index
from flip_pic_orders fpo
where fpo.flip_data_id = $1
order by fpo.pos_index