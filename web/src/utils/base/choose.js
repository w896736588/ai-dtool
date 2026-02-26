import store from '@/utils/base/store'
//只支持整数主键匹配
function ChooseDefault(business , dataList , defaultConfig ,  fieldKey){
    let cacheId = store.GetStoreIdInt(business)
    if(fieldKey === '' || fieldKey === null || fieldKey === undefined){
        fieldKey = 'id'
    }
    if(!dataList || dataList.length === 0){
        return {
            id : '-1',
            config : defaultConfig,
        }
    }
    for(let i = 0 ; i < dataList.length ; i++){
        if(dataList[i][fieldKey] && parseInt(dataList[i][fieldKey]) === parseInt(cacheId)){
            return {
                id : '' + cacheId,
                config : dataList[i],
            }
        }
    }
    //某些新增的id默认为0
    if(cacheId === -1){
        return {
            id : '-1',
            config : defaultConfig,
        }
    }
    let configTemp = dataList[0]
    return {
        id : '' + configTemp[fieldKey],
        config : configTemp,
    }
}

function ChooseId(business , chooseId){
    store.setStore(business , chooseId)
}


export default {
    ChooseDefault,
    ChooseId,
}