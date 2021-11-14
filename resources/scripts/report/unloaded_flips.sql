CREATE MATERIALIZED VIEW report.mv_unloaded_flips AS
SELECT k.epoch,
       k.total           AS participants,
       sum(k.nounloaded) AS nounloaded,
       sum(k.unloaded1)  AS unloaded1,
       sum(k.unloaded2)  AS unloaded2,
       sum(k.unloaded3)  AS unloaded3plus
FROM (SELECT p.epoch,
             CASE
                 WHEN p.unloaded_flips = 0 THEN 1
                 ELSE 0
                 END                              AS nounloaded,
             CASE
                 WHEN p.unloaded_flips = 1 THEN 1
                 ELSE 0
                 END                              AS unloaded1,
             CASE
                 WHEN p.unloaded_flips = 2 THEN 1
                 ELSE 0
                 END                              AS unloaded2,
             CASE
                 WHEN p.unloaded_flips > 2 THEN 1
                 ELSE 0
                 END                              AS unloaded3,
             count(*) OVER (PARTITION BY p.epoch) AS total
      FROM report.mv_participants p) k
GROUP BY k.epoch, k.total
ORDER BY k.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_unloaded_flips()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    call report.refresh_participants();
    refresh materialized view report.mv_unloaded_flips;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_unloaded_flips', 'report.refresh_unloaded_flips', 'e', 30, 'UnloadedFlips',
        null);