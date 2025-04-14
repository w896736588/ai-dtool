package ai_action_tpl

func ActionList() string {
	return `
    /**
     * 列表
     * User: frog
     * Date: 2024/5/16 2:34
     */
    public function actionList()
    {
        $admin_user_id = StaffServices::getAdminUserIdByStaffId(getUserId());
        $service = new TemplateService();
        $type = $this->getParam('type') ?? '';
        if(!in_array($type , [$service::TYPE_TIMEOUT_CLOSE , $service::TYPE_VISITOR_NO_RESPONSE_CLOSE])){
            format_err('类型错误');
        }
        $list = $service->getList($admin_user_id , $type);
        format_ok('' , compact('list'));
    }

`
}
