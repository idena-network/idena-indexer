SELECT f.cid,
       coalesce(fi."data", ''::bytea) icon,
       authors.address                author,
       coalesce(fw.word_1, 0)         word_id_1,
       coalesce(wd1.name, '')         word_name_1,
       coalesce(wd1.description, '')  word_desc_1,
       coalesce(fw.word_2, 0)         word_id_2,
       coalesce(wd2.name, '')         word_name_2,
       coalesce(wd2.description, '')  word_desc_2,
       coalesce(rfr.balance, 0)       balance,
       coalesce(rfr.stake, 0)         stake,
       a.grade
FROM answers a
         JOIN flips f ON f.tx_id = a.flip_tx_id AND f.grade = 1
         JOIN epoch_identities ei ON ei.address_state_id = a.ei_address_state_id AND ei.epoch = $1
         JOIN address_states s ON s.id = ei.address_state_id
         JOIN addresses ad ON ad.id = s.address_id AND lower(ad.address) = lower($2)
         JOIN transactions t ON t.id = f.tx_id
         JOIN addresses authors ON authors.id = t.from
         LEFT JOIN flip_icons fi ON fi.fd_flip_tx_id = f.tx_id
         LEFT JOIN reported_flip_rewards rfr ON rfr.address_id = s.address_id AND rfr.flip_tx_id = a.flip_tx_id
         LEFT JOIN flip_words fw ON fw.flip_tx_id = f.tx_id
         LEFT JOIN words_dictionary wd1 ON wd1.id = fw.word_1
         LEFT JOIN words_dictionary wd2 ON wd2.id = fw.word_2
WHERE NOT a.is_short