select e.epoch,
       e.validation_time,
       coalesce((select b.height
                 from blocks b
                          join block_flags bf on bf.block_height = b.height
                 where b.epoch = e.epoch
                   and bf.flag = 'FlipLotteryStarted'
                ), 0)                       firstBlockHeight,
       coalesce(es.min_score_for_invite, 0) min_score_for_invite
from epochs e
         left join epoch_summaries es on es.epoch = e.epoch - 1
where e.epoch = $1