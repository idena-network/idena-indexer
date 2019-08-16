select total_balance,
       total_stake
from epochs_detail
order by epoch desc
limit 1