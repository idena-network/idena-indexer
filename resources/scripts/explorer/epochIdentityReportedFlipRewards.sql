SELECT f.Cid,
       rfr.balance,
       rfr.stake
FROM reported_flip_rewards rfr
         JOIN flips f ON f.tx_id = rfr.flip_tx_id
WHERE rfr.epoch = $1
  AND rfr.address_id = (SELECT id FROM addresses WHERE lower(address) = lower($2))