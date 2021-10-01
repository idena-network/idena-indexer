CREATE MATERIALIZED VIEW report.mv_transactions_daily AS
SELECT date_trunc('day'::text, timezone('UTC'::text, to_timestamp(b."timestamp"::double precision))) AS date,
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
           END)                                                                                      AS other_trans
FROM indexer.transactions t
         JOIN indexer.blocks b ON b.height = t.block_height
GROUP BY (date_trunc('day'::text, timezone('UTC'::text, to_timestamp(b."timestamp"::double precision))))
ORDER BY (date_trunc('day'::text, timezone('UTC'::text, to_timestamp(b."timestamp"::double precision))));

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
