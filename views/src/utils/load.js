
function getExecTypeStatus(){
  return {
    supervisor_restart_all : false,
    supervisor_status_list : false,
    supervisor_restart : false,
    supervisor_stop : false,
    supervisor_config_show : false,
    wechat_kefu_change : false,
    wechat_kefu_status : false,
    WechatKefuChannelQrList : false,
    pull_branch_origin : false,
    git_status : false,
    change_branch : false,
    query_current_branch : false,
    restart_docker : false,
    show_compose : false,
    login_xkf :false,
    redis_search : false,
    query_vip_info : false,
    change_vip_type : false,
    redis_delete_batch : false,
    reduce_memory : false,
    stopListConsumer : false,
  }
}

export default {
  getExecTypeStatus,
}
