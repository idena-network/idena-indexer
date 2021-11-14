CREATE MATERIALIZED VIEW report.mv_fail_reasons AS
SELECT k.epoch,
       k.total,
       sum(k.success)        AS success,
       sum(k.late)           AS late,
       sum(k.wrong)          AS wrong,
       sum(k.missinganswers) AS missing,
       sum(k.lowscore)       AS lowscore
FROM (SELECT p.epoch,
             CASE
                 WHEN p.fail_reason = 0 THEN 1
                 ELSE 0
                 END                              AS success,
             CASE
                 WHEN p.fail_reason = 1 THEN 1
                 ELSE 0
                 END                              AS late,
             CASE
                 WHEN p.fail_reason = 3 THEN 1
                 ELSE 0
                 END                              AS wrong,
             CASE
                 WHEN p.fail_reason = 2 THEN 1
                 ELSE 0
                 END                              AS missinganswers,
             CASE
                 WHEN p.fail_reason = 4 THEN 1
                 ELSE 0
                 END                              AS lowscore,
             count(*) OVER (PARTITION BY p.epoch) AS total
      FROM report.mv_participants p) k
GROUP BY k.epoch, k.total
ORDER BY k.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_fail_reasons()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    call report.refresh_participants();
    refresh materialized view report.mv_fail_reasons;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_fail_reasons', 'report.refresh_fail_reasons', 'e', 30, 'ValidationFails',
        null);