package define

const RoleSystem = `system`
const RoleUser = `user`
const RoleAssistant = `assistant`

const NoCache = "no"
const StringSingleCache = "string_single"
const StringAllCache = "string_all"
const HashAdminCustomCache = "hash_admin_custom"

var CacheTypeMap = map[string]string{
	NoCache:              "无缓存",
	StringSingleCache:    "string单条缓存",
	StringAllCache:       "string多条缓存",
	HashAdminCustomCache: "hash自定义缓存",
}
