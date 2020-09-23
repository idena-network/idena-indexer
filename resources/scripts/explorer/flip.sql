select a.address                         author,
       f.size,
       b.timestamp,
       coalesce(da.name, '')             answer,
       coalesce(f.grade, 0) = 1          reported,
       coalesce(fs.wrong_words_votes, 0) wrong_words_votes,
       COALESCE(dfs.name, '')            status,
       t.hash                            tx_hash,
       b.hash                            block_hash,
       b.height                          block_height,
       b.epoch                           epoch,
       coalesce(fw.word_1, 0)            word_id_1,
       coalesce(wd1.name, '')            word_name_1,
       coalesce(wd1.description, '')     word_desc_1,
       coalesce(fw.word_2, 0)            word_id_2,
       coalesce(wd2.name, '')            word_name_2,
       coalesce(wd2.description, '')     word_desc_2,
       coalesce(fs.encrypted, false)     with_private_part,
       coalesce(f.grade, 0)              grade
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
         join addresses a on a.id = t.from
         left join flip_words fw on fw.flip_tx_id = f.tx_id
         left join words_dictionary wd1 on wd1.id = fw.word_1
         left join words_dictionary wd2 on wd2.id = fw.word_2
         left join dic_flip_statuses dfs on dfs.id = f.status
         left join dic_answers da on da.id = f.answer
         left join flip_summaries fs on fs.flip_tx_id = f.tx_id
where LOWER(f.cid) = LOWER($1)
  and f.delete_tx_id is null