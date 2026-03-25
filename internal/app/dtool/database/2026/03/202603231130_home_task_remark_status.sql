ALTER TABLE "tbl_home_task"
    ADD COLUMN "remark" TEXT NOT NULL DEFAULT '';

UPDATE "tbl_home_task"
SET "task_status" = CASE "task_status"
                        WHEN '进行中' THEN '开发中'
                        WHEN '暂停' THEN '对接中'
                        WHEN '已完成' THEN '已上线'
                        ELSE "task_status"
    END;
