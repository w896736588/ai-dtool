import base from '../base'

function StarList(type , callBack){
    base.BasePost('/api/StarList', {type : type} , callBack)
}

function StarAdd(id , name , key , value , type , callBack){
    base.BasePost('/api/StarAdd', {
        id : id,
        name : name,
        key : key,
        value : value,
        type : type,
    } , callBack)
}

function StarDel(id , callBack){
    base.BasePost('/api/StarDel', {id : id} , callBack)
}

export default {
    StarList,
    StarAdd,
    StarDel,
}