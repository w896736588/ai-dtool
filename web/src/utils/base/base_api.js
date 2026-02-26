import base from "@/utils/base";

var RealIp = ''

function Ip(data , func , boolForce){
    if(RealIp === '' || boolForce){
        base.BasePost('/api/Ip', data , function (response){
            if(response.Data && response.Data.ip){
                RealIp = response.Data.ip
                func(RealIp)
            }
        })
    }else{
        func(RealIp)
    }
}



export default {
    Ip,
}