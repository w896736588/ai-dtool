function Exist(value , arrayList){
    let boolExist = false
    for(let i in arrayList){
        if(arrayList[i] === value){
            boolExist = true
            break
        }
    }
    return boolExist
}

function SortByKey(arrayList , key , type){
    if (!arrayList || !Array.isArray(arrayList)) {
        return []
    }
    arrayList.sort((a,b) => {
        if(type.toLowerCase() === 'asc'){
            if ((a[key] || '') < (b[key] || '')) {
                return -1;
            }else{
                return 1;
            }
        }else if(type.toLowerCase() === 'desc'){
            if ((a[key] || '') < (b[key] || '')) {
                return 1;
            }else{
                return -1;
            }
        }
    })
    return arrayList
}

function MergeList(sourceList , addList){
    let targetList = []
    for(let j in sourceList){
        targetList.push(sourceList[j])
    }
    for(let i in addList){
        targetList.push(addList[i])
    }
    return targetList
}

function FilterEmpty(sourceList){
    let targetList = []
    for(let i in sourceList){
        if(sourceList[i] !== ''){
            targetList.push(sourceList[i])
        }
    }
    return targetList
}

function DeleteValueByIntKey(list , key , value){
    let newList = []
    for(let i in list){
        if(list[i][key] && parseInt(list[i][key]) !== parseInt(value)){
            newList.push(list[i])
        }
    }
    return newList
}

function DeleteValueByStringKey(list , key , value){
    let newList = []
    for(let i in list){
        if(list[i][key] && list[i][key] !== value){
            newList.push(list[i])
        }
    }
    return newList
}
export default {
    Exist,
    SortByKey,
    MergeList,
    FilterEmpty,
    DeleteValueByIntKey,
    DeleteValueByStringKey,
}