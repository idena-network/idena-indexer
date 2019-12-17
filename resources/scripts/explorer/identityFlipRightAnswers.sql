select da.name answer, count(*)
from flips f
         join transactions t on t.id = f.tx_id
         join addresses s on s.id = t.from
    -- consider only flips with answers
         join dic_answers da on da.id = f.answer
where lower(s.address) = lower($1)
group by da.name