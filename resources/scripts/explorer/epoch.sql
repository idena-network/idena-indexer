select e.validation_time,
       coalesce((select b.height
                 from blocks b
                          join block_flags bf on bf.block_height = b.height
                 where b.epoch = e.epoch
                   and bf.flag = 'FlipLotteryStarted'
                ), 0) firstBlockHeight
from epochs e
where e.epoch = $1