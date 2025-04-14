package ai_action_tpl

func ActionCreate() string {
	return `
    /**
     * 创建
     * User: frog
     * Date: 2024/5/16 2:33
     */
    public function actionCreate()
    {
        $service = new TemplateService();
        $data = BatchRequestParams();
        $ret = $service->create($this->adminUserId , $data);
        format_code($ret);
    }

`
}
