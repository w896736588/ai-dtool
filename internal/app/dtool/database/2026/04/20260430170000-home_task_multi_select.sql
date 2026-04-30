ALTER TABLE "tbl_home_task"
    ADD COLUMN "git_ids" TEXT NOT NULL DEFAULT '[]';
ALTER TABLE "tbl_home_task"
    ADD COLUMN "api_dev_entries" TEXT NOT NULL DEFAULT '[]';
