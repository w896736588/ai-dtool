function IsArray(value){
    return Array.isArray(value);
}

function IsObjectOrArray(value) {
    const type = Object.prototype.toString.call(value);
    return type === '[object Object]' || type === '[object Array]';
}

function IsObject(value) {
    return Object.prototype.toString.call(value) === '[object Object]';
}

function IsString(value) {
    return typeof value === 'string';
}

function IsNumber(value) {
    return typeof value === 'number';
}

function IsJson(value){
    try{
        let v = JSON.parse(value)
        if(IsObjectOrArray(v)){
            return true
        }else{
            return false
        }
    }catch (e){
        return false
    }
}

export default {
    IsArray,
    IsObject,
    IsString,
    IsNumber,
    IsObjectOrArray,
    IsJson,
}