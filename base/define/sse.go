package define

const (
	SseVariable   = `variable`
	SseDocker     = `docker`
	SseGit        = `git`
	SseSupervisor = `supervisor`
	SseAiCode     = `ai_code`
)

const (
	SseEventClean = `[CLEAN]`                   //清除前端的数据
	SseEventLogin = `[LOGIN_USERNAME_PASSWORD]` //通知前端弹窗输入账号密码
)
