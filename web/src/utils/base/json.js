import t from '@/utils/base/type'

function JsonDecode(str){
    let obj = JSON.parse(str)
    if(t.IsObjectOrArray(obj)){
        return obj
    }else{
        return null
    }
}
export default {
    JsonDecode,
}