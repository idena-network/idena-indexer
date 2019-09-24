select a.address,
       fr.balance,
       fr.type
from fund_rewards fr
         join blocks b on b.height = fr.block_height
         join addresses a on a.id = fr.address_id
where b.epoch = $1