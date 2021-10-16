CREATE MATERIALIZED VIEW report.mv_coins AS
SELECT n.epoch,
       CASE
           WHEN date_trunc('day'::text, n.valid_date)::date > CURRENT_DATE THEN CURRENT_DATE
           ELSE date_trunc('day'::text, n.valid_date)::date
           END                                                      AS "timestamp",
       COALESCE(n.inv1y + n.inv2y + n.dev3y + n.dev5y + n.res20y + n.res21y + n.ambas_w + n.found_w + n.prevfound_w +
                n.zero_w + n.prevdev1_w + n.prevdev2_w, 0::numeric) AS vested,
       COALESCE(n.zero_w, 0::numeric)                               AS zerow,
       COALESCE(c.minted, 0::numeric)                               AS minted,
       COALESCE(c.burnt, 0::numeric)                                AS burnt,
       COALESCE(c.staked, 0::numeric)                               AS staked,
       COALESCE(c.minted - c.burnt, 0::numeric)                     AS totalsupply
FROM (SELECT DISTINCT m.epoch,
                      m.valid_date,
                      sum(m.inv1) OVER (ORDER BY m.epoch)          AS inv1y,
                      sum(m.inv2) OVER (ORDER BY m.epoch)          AS inv2y,
                      sum(m.dev3) OVER (ORDER BY m.epoch)          AS dev3y,
                      sum(m.dev5) OVER (ORDER BY m.epoch)          AS dev5y,
                      sum(m.res20) OVER (ORDER BY m.epoch)         AS res20y,
                      sum(m.res21) OVER (ORDER BY m.epoch)         AS res21y,
                      sum(m.ambas) OVER (ORDER BY m.epoch)         AS ambas_w,
                      sum(m.found_amt) OVER (ORDER BY m.epoch)     AS found_w,
                      sum(m.prevfound_amt) OVER (ORDER BY m.epoch) AS prevfound_w,
                      sum(m.zero_amt) OVER (ORDER BY m.epoch)      AS zero_w,
                      sum(m.prevdev1_amt) OVER (ORDER BY m.epoch)  AS prevdev1_w,
                      sum(m.prevdev2_amt) OVER (ORDER BY m.epoch)  AS prevdev2_w
      FROM (SELECT ep.epoch,
                   to_timestamp(ep.validation_time::double precision) AS valid_date,
                   d.inv1_amount                                      AS inv1,
                   d.inv2_amount                                      AS inv2,
                   d.dev3_amount                                      AS dev3,
                   d.dev3_amount                                      AS dev5,
                   d.res20_amount                                     AS res20,
                   d.res21_amount                                     AS res21,
                   d.ambas_amount                                     AS ambas,
                   d.found_amount                                     AS found_amt,
                   d.prevfound_amount                                 AS prevfound_amt,
                   d.zero_amount                                      AS zero_amt,
                   d.prevdev1_amount                                  AS prevdev1_amt,
                   d.prevdev2_amount                                  AS prevdev2_amt
            FROM indexer.epochs ep
                     LEFT JOIN (SELECT bl.epoch,
                                       CASE
                                           WHEN b.address_id = 7282 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS inv1_amount,
                                       CASE
                                           WHEN b.address_id = 7283 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS inv2_amount,
                                       CASE
                                           WHEN b.address_id = 65801 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS dev3_amount,
                                       CASE
                                           WHEN b.address_id = 65800 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS dev5_amount,
                                       CASE
                                           WHEN b.address_id = 7277 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS res20_amount,
                                       CASE
                                           WHEN b.address_id = 7278 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS res21_amount,
                                       CASE
                                           WHEN b.address_id = 7276 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS ambas_amount,
                                       CASE
                                           WHEN b.address_id = 1771 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS found_amount,
                                       CASE
                                           WHEN b.address_id = 1542 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS prevfound_amount,
                                       CASE
                                           WHEN b.address_id = 940 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS zero_amount,
                                       CASE
                                           WHEN b.address_id = 7285 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS prevdev1_amount,
                                       CASE
                                           WHEN b.address_id = 7286 THEN b.balance_new - b.balance_old
                                           ELSE 0::numeric
                                           END AS prevdev2_amount
                                FROM indexer.balance_updates b
                                         JOIN indexer.blocks bl ON b.block_height = bl.height
                                WHERE b.address_id = ANY
                                      (ARRAY [7282::bigint, 7283::bigint, 65800::bigint, 65801::bigint, 7276::bigint, 7277::bigint, 7278::bigint, 1771::bigint, 1542::bigint, 940::bigint, 7285::bigint, 7286::bigint])) d
                               ON d.epoch = ep.epoch
            ORDER BY ep.epoch) m) n
         JOIN (SELECT cum.epoch,
                      cum.minted,
                      cum.burnt,
                      st.staked
               FROM (SELECT d.epoch,
                            sum(d.minted) OVER (ORDER BY d.epoch) AS minted,
                            sum(d.burnt) OVER (ORDER BY d.epoch)  AS burnt
                     FROM (SELECT b.epoch,
                                  sum(c_1.minted) AS minted,
                                  sum(c_1.burnt)  AS burnt
                           FROM indexer.coins c_1
                                    JOIN indexer.blocks b ON b.height = c_1.block_height
                           GROUP BY b.epoch) d
                     ORDER BY d.epoch) cum
                        JOIN (SELECT max_bl.epoch,
                                     c_1.total_stake AS staked
                              FROM indexer.coins c_1
                                       JOIN (SELECT bl.epoch,
                                                    max(bl.height) AS height
                                             FROM indexer.blocks bl
                                             GROUP BY bl.epoch) max_bl ON max_bl.height = c_1.block_height) st
                             ON st.epoch = cum.epoch
               ORDER BY st.epoch) c ON n.epoch = c.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_coins_daily()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    refresh materialized view report.mv_coins;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_coins', 'report.refresh_coins_daily', 'd', null, 'Coins', null);