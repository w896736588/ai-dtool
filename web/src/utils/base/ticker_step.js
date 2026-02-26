/**
 * 定时任务触发 活跃式
 */
let tickerList = [];

//主定时器
let mainTicker = null

//注册
function Register(key , step , func){
    for(let i in tickerList){
        if(tickerList[i].key === key){
            return
        }
    }
    let registerConfig = {
        key : key,
        step : step,
        triggerTime : 0,
        func : func,
        triggerNum : 0,
    }
    tickerList.push(registerConfig)
    if(mainTicker === null){
        mainTicker = setInterval(function (){
            for(let i = 0 ; i < tickerList.length ; i++){
                let item = tickerList[i]
                if(item.triggerTime <= Date.now()){
                    item.func()
                    item.triggerNum++ //递增1次
                    if(item.triggerTime > 10){ //最大10次
                        item.triggerTime = 0
                    }
                    item.triggerTime = Date.now() + item.triggerNum * item.step * 1000 //下一次出发时间
                }
            }
        } , 1000)
    }
}

//激活
function Active(key){
    for (let i = 0; i < tickerList.length; i++) {
        if(tickerList[i].key === key){
            tickerList[i].triggerTime = 0
            tickerList[i].triggerNum = 0
            break;
        }
    }
}

export default {
    Register,
    Active,
}