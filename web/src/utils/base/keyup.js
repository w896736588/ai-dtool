// 按键防抖检测类 - 优化版本
class KeyDebounceDetector {
    constructor() {
        this.downs = {}
        this.regis = []
    }
    //注册快捷键
    Register(key1 , key2 , key3 , func){
        let keys = []
        if(key1 !== undefined && key1 !== null && key1 !== ''){
            keys.push(key1)
        }
        if(key2 !== undefined && key2 !== null && key2 !== ''){
            keys.push(key2)
        }
        if(key3 !== undefined && key3 !== null && key3 !== ''){
            keys.push(key3)
        }
        if(keys.length > 0){
            this.regis.push({
                keys : keys,
                func : func
            })
        }
    }
    // 手动添加按键释放
    keyUp(key) {
        this.downs[key] = 0
    }
    //按下键
    keyDown(key) {
        if(this.downs[key] && this.downs[key] > 0){
            return
        }
        this.downs[key] = Date.now()
        let sortedKeys = Object.keys(this.downs)
            .filter(key => this.downs[key] !== 0)
            .sort((a, b) => this.downs[a] - this.downs[b]);
        console.log('当前触发的' , sortedKeys)
        for(let i in this.regis){
            let trigger = true
            let register = this.regis[i]
            if(sortedKeys.length === register.keys.length){
                for (let j in register.keys) {
                    if(register.keys[j] !== sortedKeys[j]){
                        trigger = false
                    }
                }
            }else{
                trigger = false
            }
            if(trigger){
                register.func()
            }
        }
    }

}

export default KeyDebounceDetector;