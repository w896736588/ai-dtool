import json from "@/utils/base/json";
import t from "@/utils/base/type";

//最多支持2级对象
function Explain(sourceData) {
    let returnObj = getReturnData(sourceData)
    //准备返回的数据
    let returnTemplate = returnObj.returnTemplate
    let returnConfig = returnObj.returnConfig
    let returnTemplateObj = returnObj.returnTemplateObj
    let returnConfigData = null;

    //解析 returnTemplateObj [{"label":"","value":"","link":"","userList":[{"user_name":"","password":""}]}]
    if (t.IsArray(returnTemplateObj)) { //数组解析
        returnConfigData = []
        returnConfigData = explainArray(returnConfig , returnTemplateObj , returnConfigData);
    } else { //对象解析
        returnConfigData = {}
        returnConfigData = explainObj(returnConfig , returnTemplateObj , returnConfigData)
    }
    return {returnConfigData: returnConfigData , returnTemplate : returnTemplate , returnTemplateObj : returnTemplateObj}
}

//解析数组
function explainArray(returnConfig , returnTemplateObj , returnConfigList){
    if(returnTemplateObj.length === 0){
        return [];
    }
    let returnTemplateObjT = returnTemplateObj[0] //取第一个值就行
    returnConfig.forEach((value1, index1) => {
        let tempConfig = {}
        console.log(returnTemplateObjT)
        Object.keys(returnTemplateObjT).forEach(function(index2) {
            let value2 = returnTemplateObjT[index2]
            if(t.IsArray(value2)){ //非字符串的 再次进行解析
                let tempSourceData = {
                    template : value2 ,
                    config : value1[index2]
                }
                let explainResult = Explain(tempSourceData)
                tempConfig[index2] = explainResult.returnConfigData
            }else if(t.IsObject(value2)){ //非字符串的 再次进行解析
                let tempSourceData = {
                    template : value2 ,
                    config : value1[index2]
                }
                let explainResult = Explain(tempSourceData)
                tempConfig[index2] = explainResult.returnConfigData
            }else if (value1[index2]) {
                tempConfig[index2] = value1[index2]
            } else {
                tempConfig[index2] = ''
            }
        });
        returnConfigList.push(tempConfig)
    });
    return returnConfigList
}

//解析对象
function explainObj(returnConfig , returnTemplateObj , returnConfigObj){
    let tempConfig = {}
    Object.keys(returnTemplateObj).forEach(function(index2){
        let value2 = returnTemplateObj[index2]
        if(t.IsArray(value2)){ //非字符串的 再次进行解析
            let tempSourceData = {
                template : value2 ,
                config : returnConfig[index2]
            }
            let explainResult = Explain(tempSourceData)
            tempConfig[index2] = explainResult.returnConfigData
        }else if(t.IsObject(value2)){ //非字符串的 再次进行解析
            let tempSourceData = {
                template : value2 ,
                config : returnConfig[index2]
            }
            let explainResult = Explain(tempSourceData)
            tempConfig[index2] = explainResult.returnConfigData
        }else if (returnConfig[index2]) {
            tempConfig[index2] = returnConfig[index2]
        } else {
            tempConfig[index2] = ''
        }
    });
    returnConfigObj = tempConfig
    return returnConfigObj
}

function getReturnData(sourceData) {
    let returnTemplate = null //模板原始内容
    let returnConfig = null //配置的值
    let returnTemplateObj = null //模板对象

    let objSourceData = null
    if (t.IsString(sourceData)) {
        objSourceData = json.JsonDecode(sourceData)
    } else {
        objSourceData = sourceData
    }

    //模板
    if (objSourceData.template) {
        returnTemplate = objSourceData.template
        if(t.IsString(returnTemplate)){
            returnTemplateObj = json.JsonDecode(returnTemplate)
        }else{
            returnTemplateObj = returnTemplate
            returnTemplate = JSON.stringify(returnTemplateObj)
        }
    } else {
        returnTemplate = ''
        returnConfig = []
        returnTemplateObj = []
        return {returnTemplate: returnTemplate, returnTemplateObj: returnTemplateObj, returnConfig: returnConfig}
    }
    //配置
    if (objSourceData.config) {
        returnConfig = objSourceData.config
        if (t.IsArray(returnTemplateObj)) {
            if (t.IsString(returnConfig)) { //如果是字符串 那么久转
                returnConfig = json.JsonDecode(returnConfig)
            }else if (!t.IsArray(returnConfig)) { //如果非数组 那么设置为空数组
                returnConfig = []
            }
        }else if(t.IsObject(returnTemplateObj)){
            if (t.IsString(returnConfig)) { //如果是字符串 那么久转
                returnConfig = json.JsonDecode(returnConfig)
            }else if (!t.IsObject(returnConfig)) { //如果非数组 那么设置为空数组
                returnConfig = []
            }
        }
    } else {
        if (t.IsArray(returnTemplateObj)) {
            returnConfig = []
        } else {
            returnConfig = {}
        }
    }
    return {returnTemplate: returnTemplate, returnTemplateObj: returnTemplateObj, returnConfig: returnConfig}
}

export default {
    Explain,
}