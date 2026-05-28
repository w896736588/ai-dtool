ALTER TABLE "tbl_task_workflow_chat" RENAME TO "agent_chat";
ALTER TABLE "agent_chat" RENAME COLUMN "workflow_id" TO "from_id";
ALTER TABLE "agent_chat" ADD COLUMN "from_type" TEXT NOT NULL DEFAULT '';
UPDATE "agent_chat" SET "from_type" = 'work_flow' WHERE "from_type" = '';
CREATE INDEX "idx_agent_chat_from" ON "agent_chat" ("from_type", "from_id", "id" DESC);
CREATE INDEX "idx_agent_chat_agent_cli" ON "agent_chat" ("agent_cli_id", "id" DESC);
