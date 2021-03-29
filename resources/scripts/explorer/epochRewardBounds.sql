SELECT erb.bound_type, erb.min_amount, mina.address min_address, erb.max_amount, maxa.address max_address
FROM epoch_reward_bounds erb
         JOIN addresses mina on mina.id = erb.min_address_id
         JOIN addresses maxa on maxa.id = erb.max_address_id
WHERE erb.epoch = $1
ORDER BY erb.bound_type