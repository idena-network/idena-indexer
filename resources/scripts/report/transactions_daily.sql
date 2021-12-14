CREATE MATERIALIZED VIEW report.mv_transactions_daily AS
SELECT k.date,
       k.invite_trans,
       k.send_trans,
       k.flip_trans,
       k.valid_trans,
       k.contract_trans,
       k.deleg_trans,
       k.mining_trans,
       k.other_trans,
       sum(k.invite_trans) OVER (ORDER BY k.date)                 AS cum_invite_trans,
       sum(k.send_trans) OVER (ORDER BY k.date)                   AS cum_send_trans,
       sum(k.flip_trans) OVER (ORDER BY k.date)                   AS cum_flip_trans,
       sum(k.valid_trans) OVER (ORDER BY k.date)                  AS cum_valid_trans,
       sum(k.contract_trans) OVER (ORDER BY k.date)               AS cum_contract_trans,
       sum(k.deleg_trans) OVER (ORDER BY k.date)                  AS cum_deleg_trans,
       sum(k.mining_trans) OVER (ORDER BY k.date)                 AS cum_mining_trans,
       sum(k.other_trans) OVER (ORDER BY k.date)                  AS cum_other_trans,
       sum(k.invite_trans + k.send_trans + k.flip_trans + k.valid_trans + k.contract_trans + k.deleg_trans +
           k.mining_trans + k.other_trans) OVER (ORDER BY k.date) AS cum_trans,
       k.fee,
       sum(k.fee) OVER (ORDER BY k.date)                          AS cum_fee
FROM (SELECT date_trunc('day'::text, timezone('UTC'::text, to_timestamp(b."timestamp"::double precision))) AS date,
             sum(CASE
                     WHEN t.type IN (1, 2, 10) THEN 1
                     ELSE 0
                 END)                                                                                      AS invite_trans,
             sum(CASE
                     WHEN t.type = 0 THEN 1
                     ELSE 0
                 END)                                                                                      AS send_trans,
             sum(CASE
                     WHEN t.type IN (4, 14) THEN 1
                     ELSE 0
                 END)                                                                                      AS flip_trans,
             sum(CASE
                     WHEN t.type IN (5, 6, 7, 8) THEN 1
                     ELSE 0
                 END)                                                                                      AS valid_trans,
             sum(CASE
                     WHEN t.type IN (15, 16, 17) THEN 1
                     ELSE 0
                 END)                                                                                      AS contract_trans,
             sum(CASE
                     WHEN t.type IN (18, 19, 20) THEN 1
                     ELSE 0
                 END)                                                                                      AS deleg_trans,
             sum(CASE
                     WHEN t.type = 9 THEN 1
                     ELSE 0
                 END)                                                                                      AS mining_trans,
             sum(CASE
                     WHEN t.type IN (21, 11, 12, 13, 3) THEN 1
                     ELSE 0
                 END)                                                                                      AS other_trans,
             sum(t.fee)                                                                                    AS fee
      FROM indexer.transactions t
               JOIN indexer.blocks b ON b.height = t.block_height
      GROUP BY (date_trunc('day'::text, timezone('UTC'::text, to_timestamp(b."timestamp"::double precision))))) k
ORDER BY k.date;

CREATE OR REPLACE PROCEDURE report.refresh_txs_daily()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    refresh materialized view report.mv_transactions_daily;
END
$$;


INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_transactions_daily', 'report.refresh_txs_daily', 'd', null, 'TransactionsDaily', null);
