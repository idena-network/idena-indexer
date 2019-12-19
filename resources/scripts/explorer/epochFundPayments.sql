select a.address,
       fr.balance,
       dert.name "type"
from fund_rewards fr
         join blocks b on b.height = fr.block_height
         join addresses a on a.id = fr.address_id
         join dic_epoch_reward_types dert on dert.id = fr.type
where b.epoch = $1