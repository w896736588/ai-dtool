package define

const ErrorCodeSuccess = 0
const ErrorCodeErrorUniKey = 1
const ErrorCodeRunError = 2
const ErrorKeyNotExist = 3
const ErrorKeyExist = 4

const CacheString = `string`
const CacheHash = `hash`
const CacheList = `list`
const CacheSet = `set`
const CacheZSet = `zset`

const ENTER = `
`

var VipMap = map[string]string{
	`0`: `免费版`,
	`1`: `专业版`,
	`2`: `企业版`,
	`3`: `标准版`,
	`4`: `平台版`,
}
