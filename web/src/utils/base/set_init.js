import base from "@/utils/base";
import store from "@/utils/base/store";
//设置组件需要重新初始化
function SetIsInit(componentName){
    store.setStore(componentName + '_is_init' , 1)
}

function GetIsInit(componentName){
    let isInit = store.getStore(componentName + '_is_init')
    return parseInt(isInit) === 1
}

function DelInit(componentName){
    store.removeStore(componentName + '_is_init')
}


export default {
    SetIsInit,
    GetIsInit,
    DelInit,
}