package worker

// ToolDefinitions 返回 OpenAI 格式的工具定义列表，供 FC 循环使用。
func ToolDefinitions() []map[string]any {
	return []map[string]any{
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolFileRead,
				`description`: `读取文件内容。返回文件的完整文本内容。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `要读取的文件路径`,
						},
					},
					`required`: []string{`path`},
				},
			},
		},

		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolHttpCall,
				`description`: `调用 dtool 的 HTTP API 接口。所有接口均为 POST 方法，基地址已自动拼接，只需传接口路径和 JSON 请求体。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `API 接口路径，如 /api/GitConfigList、/api/GitRemoteBranchList`,
						},
						`body`: map[string]any{
							`type`:        `string`,
							`description`: `JSON 格式的请求体，如 {}、{"ssh_id":"5","code_path":"/var/www/common3"}`,
						},
					},
					`required`: []string{`path`, `body`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolRunScript,
				`description`: `【严格限制】仅用于执行 skills/ 目录下的预置脚本。绝对禁止用它执行新建脚本——所有查询和操作一律使用 http_call 调 API。仅当步骤文件明确指示调用某个预置脚本时才可使用此工具。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `预置脚本路径（仅限 skills/ 下已有脚本），如 skills/dtool-git/scripts/git_api.py`,
						},
						`args`: map[string]any{
							`type`:        `string`,
							`description`: `命令行参数（空格分隔），如 --repo_name common3 --branch develop`,
						},
						`timeout`: map[string]any{
							`type`:        `string`,
							`description`: `超时秒数，默认 60 秒`,
						},
					},
					`required`: []string{`path`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolAskUser,
				`description`: `向用户提问确认，暂停当前任务等待用户回复。仅当缺少必要信息（操作对象不明确、参数不足）或需要确认危险操作时使用。只读查询无需确认。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`question`: map[string]any{
							`type`:        `string`,
							`description`: `向用户提问的内容，应清晰列出选项或需要补充的信息`,
						},
						`options`: map[string]any{
							`type`:        `string`,
							`description`: `可选选项列表，用逗号分隔，如 common3-web,common3-api,common3-admin。为空则用户自由回答`,
						},
						`reason`: map[string]any{
							`type`:        `string`,
							`description`: `需要确认的原因，如 操作对象不明确、危险操作确认、参数不足`,
						},
					},
					`required`: []string{`question`, `reason`},
				},
			},
		},
	}
}
