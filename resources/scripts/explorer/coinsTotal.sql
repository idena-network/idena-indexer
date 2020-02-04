SELECT c.total_balance,
       c.total_stake
FROM coins c
ORDER BY c.block_height DESC
LIMIT 1