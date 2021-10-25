CREATE MATERIALIZED VIEW report.mv_invite_activations
AS
SELECT k."timestamp",
       CASE
           WHEN k."timestamp" = CURRENT_TIMESTAMP AND k.length <> 0 THEN
                   (((k.length - 1) * 24)::double precision + k.hour_length) / (k.epoch_length * 24)::real
           ELSE k.length::double precision / k.epoch_length::real
           END AS share_of_period,
       k.activations,
       k.cum_activations,
       k.epoch
FROM (SELECT c."timestamp",
             c.start_date,
             date_part('days'::text, c."timestamp" - c.start_date::timestamp with time zone)::integer AS length,
             date_part('hour'::text, c."timestamp" - c.start_date::timestamp with time zone)          AS hour_length,
             date_part('days'::text, c.finish_date - c.start_date)::integer                           AS epoch_length,
             c.activations,
             c.cum_activations,
             c.epoch
      FROM (SELECT CASE
                       WHEN b.date = b.finish_date THEN b.date + '13:30:00'::time without time zone::interval
                       WHEN b.date = CURRENT_DATE THEN CURRENT_TIMESTAMP
                       WHEN b.date = b.start_date THEN b.date + '13:30:00'::time without time zone::interval
                       ELSE b.date + '23:59:59'::time without time zone::interval
                       END                                          AS "timestamp",
                   b.epoch,
                   b.start_date::text::timestamp without time zone  AS start_date,
                   b.finish_date::text::timestamp without time zone AS finish_date,
                   b.cum_activations,
                   b.activations
            FROM (SELECT v.date,
                         first_value(v.date)
                         OVER (PARTITION BY v.epoch ORDER BY v.date RANGE BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) AS start_date,
                         date_trunc('day'::text,
                                    to_timestamp(ep.validation_time::double precision))                                        AS finish_date,
                         v.epoch,
                         v.activations,
                         sum(v.activations) OVER (PARTITION BY v.epoch ORDER BY v.date)                                        AS cum_activations
                  FROM (SELECT date_trunc('day'::text, to_timestamp(b_1."timestamp"::double precision)) AS date,
                               b_1.epoch,
                               sum(CASE
                                       WHEN t_1.type = 2 THEN 1
                                       ELSE 0
                                   END)                                                                 AS invites,
                               sum(CASE
                                       WHEN t_1.type = 1 THEN 1
                                       ELSE 0
                                   END)                                                                 AS activations
                        FROM indexer.transactions t_1
                                 JOIN indexer.blocks b_1 ON b_1.height = t_1.block_height
                        WHERE t_1.type = ANY (ARRAY [1, 2])
                          AND b_1.epoch >= ((SELECT max(epoch) AS max FROM indexer.epochs) - 2)
                        GROUP BY b_1.epoch, date_trunc('day'::text, to_timestamp(b_1."timestamp"::double precision))
                       ) v
                           JOIN indexer.epochs ep ON ep.epoch = v.epoch
                 ) b) c) k;

CREATE OR REPLACE PROCEDURE report.refresh_invites_daily()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    refresh materialized view report.mv_invite_activations;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_invite_activations', 'report.refresh_invites_daily', 'd', null, 'InviteActivations', null);