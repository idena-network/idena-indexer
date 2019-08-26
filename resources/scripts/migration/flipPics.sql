select fp.data
from flip_pics fp
where fp.flip_data_id = $1
order by fp.index