select count(*) block_count
from blocks b
where b.epoch = $1