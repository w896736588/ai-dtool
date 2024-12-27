package base

type Env struct {
	IsBuild  bool   //true 打包模式 false go run模式
	RootPath string //项目根目录
	AppName  string //项目名称
}
