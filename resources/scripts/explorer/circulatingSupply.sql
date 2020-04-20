SELECT c.total_balance - coalesce(excluded.balance, 0) total_balance,
       c.total_stake - coalesce(excluded.stake, 0)     total_stake
FROM coins c,
     (
         select sum(balance) balance, sum(stake) stake
         from balances
         where address_id in (select id from addresses where lower(address) = any ($1::text[]))
     ) excluded
ORDER BY c.block_height DESC
LIMIT 1