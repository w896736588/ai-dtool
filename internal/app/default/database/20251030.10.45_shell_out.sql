ALTER TABLE tbl_shell_out ADD COLUMN regex_errors text NOT NULL DEFAULT '';
ALTER TABLE tbl_shell_out ADD COLUMN regex_no_errors text NOT NULL DEFAULT '';
ALTER TABLE tbl_shell_out ADD COLUMN regex_filters NOT NULL DEFAULT '';