package ai_action_tpl

func ActionDelete() string {
	return `
    /**
     * 删除
     * User: frog
     * Date: 2024/5/16 2:34
     */
    public function actionDelete()
    {
        $service = new TemplateService();
        $id = $this->getParam('id') ?: '';
        if(empty($id)){
            format_err('ID不能为空');
        }
        $ret = $service->delete($this->adminUserId , $id);
        format_code($ret);
    }

`
}
