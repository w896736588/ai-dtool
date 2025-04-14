package ai_action_tpl

func ActionDetail() string {
	return `
    /**
     * 单个明细
     * User: frog
     * Date: 2024/5/16 2:33
     */
    public function actionDetail()
    {
        $service = new TemplateService();
        $id = $this->getParam('id') ?? '';
        if(empty($id)){
            format_err('ID不能为空');
        }
        $ret = $service->getDetail($this->adminUserId , $id);
        format_code($ret);
    }

`
}
