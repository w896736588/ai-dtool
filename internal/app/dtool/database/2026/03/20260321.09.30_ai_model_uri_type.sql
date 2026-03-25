ALTER TABLE tbl_ai_model ADD COLUMN uri TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_ai_model ADD COLUMN model_type TEXT NOT NULL DEFAULT 'llm';

UPDATE tbl_ai_model
SET uri = '/v1/chat/completions'
WHERE IFNULL(TRIM(uri), '') = '';

UPDATE tbl_ai_model
SET model_type = 'llm'
WHERE IFNULL(TRIM(model_type), '') = '';

UPDATE tbl_ai_provider
SET base_url = REPLACE(base_url, '/v1/chat/completions', '')
WHERE base_url LIKE '%/v1/chat/completions';

UPDATE tbl_ai_provider
SET base_url = REPLACE(base_url, '/chat/completions', '')
WHERE base_url LIKE '%/chat/completions';
