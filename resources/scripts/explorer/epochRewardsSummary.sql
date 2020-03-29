select b.epoch,
       tr.total,
       tr.validation,
       tr.flips,
       tr.invitations,
       tr.foundation,
       tr.zero_wallet,
       tr.validation_share,
       tr.flips_share,
       tr.invitations_share
from total_rewards tr
         join blocks b on b.height = tr.block_height
where b.epoch = $1