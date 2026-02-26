import t from "../base/type"
function SearchListObj(listSource , searchKey ) {
    let searchNum = 0
    let filterValueList = searchKey.split(' ') //支持多个条件搜索 空格
    if(!t.IsArray(listSource)){
        return [searchNum , listSource]
    }
    let list = JSON.parse(JSON.stringify(listSource))
    if(searchKey === ''){
        for (let i in list) {
            list[i]['show'] = true
            for (let j in list[i]) {
                if(t.IsString(list[i][j])){
                    list[i][j] = reset(list[i][j])
                }
            }
        }
        return [searchNum,list]
    }
    for (let i in list) {
        if(t.IsObject(list[i])){
            list[i]['show'] = false
            for (let j in list[i]) {
                if(t.IsString(list[i][j])){
                    list[i][j] = reset(list[i][j])
                    let newRet = replace(list[i][j] , filterValueList)
                    if(!newRet[1]){
                        searchNum++
                        list[i][j] = newRet[0]
                        list[i]['show'] = true
                    }
                }
            }
        }
    }
    return [searchNum , list]
}

function reset(value){
    return value.replace(/<span[^>]*>(.*?)<\/span>/g, function(match, group1) {
        return group1;
    });
}

function replace(value , filterList){
    let oldValue = value
    for(let i in filterList){
        if(filterList[i] === ''){
            continue;
        }
        let regex = new RegExp(filterList[i], 'g');
        value = value.replace(regex, '<span style="color:red">'+filterList[i]+'</span>');
    }
    return [value,oldValue===value]
}

function SearchObj(){

}

export default {
    SearchListObj,
    SearchObj,
}