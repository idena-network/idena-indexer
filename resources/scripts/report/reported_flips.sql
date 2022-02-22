CREATE MATERIALIZED VIEW report.mv_reported_flips AS
SELECT b.epoch + 1 AS epoch,
       sum(CASE
               WHEN fl.grade = 1 THEN 1
               ELSE 0
           END)    AS num_of_reports,
       count(*)    AS total_num
FROM indexer.flips fl
         LEFT JOIN indexer.transactions tr ON fl.tx_id = tr.id
         LEFT JOIN indexer.blocks b ON b.height = tr.block_height
WHERE fl.status <> 3
GROUP BY b.epoch
ORDER BY b.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_reported_flips()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    refresh materialized view report.mv_reported_flips;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_reported_flips', 'report.refresh_reported_flips', 'e', 30, 'ReportedFlips', null);
