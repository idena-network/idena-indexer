SELECT block_height, votes
FROM upgrade_voting_history
WHERE upgrade = $1
ORDER BY block_height