ALTER TABLE "agent_chat" ADD COLUMN "is_read" INTEGER NOT NULL DEFAULT 1;

UPDATE "agent_chat"
SET "is_read" = CASE
    WHEN "status" = 'running' THEN 1
    ELSE 0
END;

CREATE INDEX "idx_agent_chat_read_scope"
    ON "agent_chat" ("from_type", "from_id", "is_read", "id" DESC);

CREATE INDEX "idx_agent_chat_read_agent_cli"
    ON "agent_chat" ("agent_cli_id", "is_read", "id" DESC);
