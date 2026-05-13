ALTER TABLE tbl_ssh ADD COLUMN "post_connect_cmds" TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_ssh ADD COLUMN "cmd_timeout" integer NOT NULL DEFAULT 3;
