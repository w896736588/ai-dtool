//列表过滤
function QuickSearch(filterValue , searchList , compareKeyList){
    if(filterValue === ``){
        return {
            list : searchList,
            searchNum : searchList.length
        }
    }
    filterValue = filterValue.toLowerCase()
    let searchNum = 0
    let keyResultTemp = []
    let filterValueList = filterValue.split(' ')
    for (let i in searchList) {
        let boolShow = true
        for (let j in filterValueList) {
            if (filterValueList[j] !== '') {
                //都找不到就过滤
                let findNum = 0
                for(let compareKey in compareKeyList){
                    let findStr = searchList[i][compareKeyList[compareKey]].toLowerCase()
                    let indexKey = findStr.indexOf(filterValueList[j].toLowerCase())
                    if(indexKey >= 0) {
                        findNum++
                        break
                    }
                }
                if(findNum === 0){ //找不到直接退出
                    boolShow = false
                    break
                }
            }
        }
        if(boolShow){
            keyResultTemp.push(searchList[i])
            searchNum++
        }
    }
    return {
        list : keyResultTemp,
        searchNum : searchNum
    }
}

function updateSetValue(list , key , value){
    let boolFind = false
    let copyValue = {...value}
    for (let i in list) {
        if (list[i][key] === copyValue[key]) {
            list[i] = copyValue
            boolFind = true
        }
    }
    if(!boolFind){
        list.push(copyValue)
    }
    return list
}

function SearchSetValue(list , key , value){
    for (let i in list) {
        if (list[i][key] === value[key]) {
            return list[i]
        }
    }
    return {}
}


export default {
    QuickSearch,
    updateSetValue,
    SearchSetValue,
}