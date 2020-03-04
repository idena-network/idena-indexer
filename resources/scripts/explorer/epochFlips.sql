select f.Cid,
       f.Size,
       a.address                       author,
       b.epoch,
       COALESCE(dfs.name, '')          status,
       COALESCE(da.name, '')           answer,
       COALESCE(f.wrong_words, false)  wrongWords,
       COALESCE(ww.cnt, 0)             wrong_words_votes,
       COALESCE(short.answers, 0),
       COALESCE(long.answers, 0),
       b.timestamp,
       coalesce(fi."data", ''::bytea)  icon,
       coalesce(fw.word_1, 0)          word_id_1,
       coalesce(wd1.name, '')          word_name_1,
       coalesce(wd1.description, '')   word_desc_1,
       coalesce(fw.word_2, 0)          word_id_2,
       coalesce(wd2.name, '')          word_name_2,
       coalesce(wd2.description, '')   word_desc_2,
       coalesce(pics_count.cnt, 0) = 2 with_private_part
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from
         join blocks b on b.height = t.block_height and b.epoch = $1
         left join (select a.flip_tx_id, count(*) answers
                    from answers a
                    where a.is_short = true
                    group by a.flip_tx_id) short
                   on short.flip_tx_id = f.tx_id
         left join (select a.flip_tx_id, count(*) answers
                    from answers a
                    where a.is_short = false
                    group by a.flip_tx_id) long
                   on long.flip_tx_id = f.tx_id
         left join (select a.flip_tx_id, count(*) cnt
                    from answers a
                    where not a.is_short
                      and a.wrong_words
                    group by a.flip_tx_id) ww
                   on ww.flip_tx_id = f.tx_id
         left join flip_icons fi on fi.fd_flip_tx_id = f.tx_id
         left join flip_words fw on fw.flip_tx_id = f.tx_id
         left join words_dictionary wd1 on wd1.id = fw.word_1
         left join words_dictionary wd2 on wd2.id = fw.word_2
         left join dic_flip_statuses dfs on dfs.id = f.status
         left join dic_answers da on da.id = f.answer
         left join (select fd_flip_tx_id, count(*) cnt
                    from flip_pics
                    group by fd_flip_tx_id) pics_count on pics_count.fd_flip_tx_id = f.tx_id
where f.delete_tx_id is null
order by f.tx_id desc
limit $3
offset
$2