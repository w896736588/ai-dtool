import mod from '../module'
import copy from './copy'

//格式化字符串
function formatResult(formatString , formatTypeList){
    let formatList = mod.GetFormatList()
    for(let i = 0 ;i < formatTypeList.length ; i++){
        switch (formatTypeList[i]){
            case 'copy':
                formatString = formatCopy(formatString , formatList)
                break;
            case 'color':
                formatString = formatColor(formatString , formatList)
                break;
            case 'length':
                formatString = formatLength(formatString , formatList)
                break;
            case 'replace':
                formatString = formatReplace(formatString , formatList)
                break;
        }
    }

    return formatString
}

//格式化copy
function formatCopy(formatString , formatList){
    if(!formatList.hasOwnProperty('copy')){
        return formatString
    }
    let formatCopyList = formatList.copy
    if (formatCopyList === undefined || formatString === '' || !Array.isArray(formatCopyList)) {
        return formatString
    }
    for(let i = 0 ; i < formatCopyList.length ; i++){
        let formatVal = formatCopyList[i]
        let regex = new RegExp(formatVal.reg, formatVal.model);
        //formatString = formatString.replace(regex, '<a onclick="handleCopy(\'$&\')" style="' + formatVal.style +'" class="'+formatVal.class+'">$&</a>');
        // 使用模板字符串来避免引号问题，并且使用适当的转义
        formatString = formatString.replace(regex, (match) => replaceWithLink(match, formatVal));
    }
    return formatString
}

const isAllSpaces = str => /^\s*$/.test(str);

function  replaceWithLink(match , formatVal) {
    let rnIndex = match.indexOf('\r\n')
    let nIndex = match.indexOf('\n')
    let copyValue = ''
    let lastValue = ''
    if(rnIndex !== -1){
        copyValue = match.substring(0 , rnIndex)
        lastValue = match.substring(rnIndex)
    }else if (nIndex !== -1){
        copyValue = match.substring(0 , nIndex)
        lastValue = match.substring(nIndex)
    }else{
        copyValue = match
    }
    if(copyValue === ''){
        return copyValue
    }
    if(isAllSpaces(copyValue)){
        return copyValue
    }
    let copyIndex = copy.SetCopyContent(copyValue)
    return '<a onclick="handleCopy(' + copyIndex + ')" style="' + formatVal.style +'" class="'+formatVal.class+'">'+copyValue+'</a>' + lastValue;
}


//格式化color
function formatColor(formatString , formatList){
    if(!formatList.hasOwnProperty('color')){
        return formatString
    }
    let formatColorList = formatList.color
    if (formatColorList === undefined || formatString === '' || !Array.isArray(formatColorList)) {
        return formatString
    }
    for(let i = 0 ; i < formatColorList.length ; i++){
        let formatVal = formatColorList[i]
        let regex = new RegExp(formatVal.reg, formatVal.model);
        formatString = formatString.replace(regex, '<span style="' + formatVal.style +'" class="'+formatVal.class+'">$&</span>');
    }
    return formatString
}


//格式化copy
function formatReplace(formatString , formatList){
    if(!formatList.hasOwnProperty('replace')){
        return formatString
    }
    let formatReplaceList = formatList.replace
    if (formatReplaceList === undefined || formatString === '' || !Array.isArray(formatReplaceList)) {
        return formatString
    }
    for(let i = 0 ; i < formatReplaceList.length ; i++){
        let formatVal = formatReplaceList[i]
        let regex = new RegExp(formatVal.reg, formatVal.model);
        formatString = formatString.replace(regex, formatVal.val);
    }
    return formatString
}

//格式化长度
function formatLength(formatString , formatList){
    if(!formatList.hasOwnProperty('length')){
        return formatString
    }
    let formatLength = formatList.length
    let sourceLength = formatString.length
    return formatString.substring(sourceLength - formatLength)
}

export default {
    formatResult,
    formatCopy,
    formatReplace,
    formatLength,
}