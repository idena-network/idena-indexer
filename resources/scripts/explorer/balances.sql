SELECT b.address_id, a.address, b.balance, b.stake
FROM balances b
         JOIN addresses a ON a.id = b.address_id
WHERE $2::bigint IS NULL
   OR b.balance <= $3 AND b.address_id >= $2
ORDER BY b.balance DESC, b.address_id
LIMIT $1