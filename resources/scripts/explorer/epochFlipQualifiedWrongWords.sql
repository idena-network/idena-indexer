select f.wrong_words, count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height and b.epoch = $1
where f.delete_tx_id is null
group by f.wrong_words