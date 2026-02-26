import t from "@/utils/base/type"
//将targetObj中的属性复制到sourceObj
function Copy(sourceObj,targetObj){
    if(t.IsArray(sourceObj)){
        for (let key in sourceObj) {
            if(!targetObj[key]){
                continue
            }
            if(t.IsArray(sourceObj[key]) || t.IsObject(sourceObj[key])){
                Copy(sourceObj[key] , targetObj[key])
            }else{
                sourceObj[key] = targetObj[key]
            }
        }
    }else if(t.IsObject(sourceObj)){
        let keys = Object.keys(sourceObj);
        for (let key of keys) {
            if(!targetObj[key]){
                continue
            }
            if(t.IsArray(sourceObj[key]) || t.IsObject(sourceObj[key])){
                Copy(sourceObj[key] , targetObj[key])
            }else{
                sourceObj[key] = targetObj[key]
            }
        }
    }
}

function ToEmpty(sourceObj){
    if(t.IsArray(sourceObj)){
        for (let key in sourceObj) {
            if(t.IsArray(sourceObj[key]) || t.IsObject(sourceObj[key])){
                ToEmpty(sourceObj[key])
            }else{
                if(t.IsNumber( sourceObj[key])){
                    sourceObj[key] = 0
                }else{
                    sourceObj[key] = ''
                }
            }
        }
    }else if(t.IsObject(sourceObj)){
        let keys = Object.keys(sourceObj);
        for (let key of keys) {
            if(t.IsArray(sourceObj[key]) || t.IsObject(sourceObj[key])){
                ToEmpty(sourceObj[key])
            }else{
                if(t.IsNumber( sourceObj[key])){
                    sourceObj[key] = 0
                }else{
                    sourceObj[key] = ''
                }
            }
        }
    }
}
export default {
    Copy,
    ToEmpty,
}