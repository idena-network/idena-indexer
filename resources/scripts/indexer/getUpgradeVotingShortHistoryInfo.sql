SELECT coalesce(t1.items, 0), coalesce(t2.last_height, 0), coalesce(t2.last_step, 0)
FROM (SELECT 1) t
         LEFT JOIN (SELECT items FROM upgrade_voting_history_summary WHERE upgrade = $1) t1 ON true
         LEFT JOIN (SELECT last_height, last_step FROM upgrade_voting_short_history_summary WHERE upgrade = $1) t2
                   ON true