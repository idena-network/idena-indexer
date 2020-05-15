SELECT c.total_balance - coalesce(excluded.balance, 0) total_balance
FROM coins c,
     (
         select sum(balance) balance
         from balances
         where address_id in (select id from addresses where lower(address) = any ($1::text[]))
     ) excluded
ORDER BY c.block_height DESC
LIMIT 1