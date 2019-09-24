select b.epoch,
       tr.total,
       tr.validation,
       tr.flips,
       tr.invitations,
       tr.foundation,
       tr.zero_wallet
from total_rewards tr
         join blocks b on b.height = tr.block_height
where b.epoch = $1