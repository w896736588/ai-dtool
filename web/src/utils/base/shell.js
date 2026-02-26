import base from "@/utils/base";

var lastScrollTime = 0
//自动计算并设置mainCard最底部离屏幕底部的高度
function calculateShellDivHeight(_that) {
    setTimeout(function (){
        let mainCard = document.getElementById('mainCard')
        if(!mainCard){
            return
        }
        let rect = mainCard.getBoundingClientRect();
        let viewportHeight = window.innerHeight || document.documentElement.clientHeight;
        _that.shellController.divHeight = viewportHeight - rect.bottom - 15;
        //自动置位最底下
        // ShellDivToBottom(true)
    } , 500)
}

//自动滚动到最底下
function ShellDivToBottom(force){
    if(!force && Date.now() - lastScrollTime < 200){
        return
    }

    lastScrollTime = Date.now()
    setTimeout(function(){
        const scrollbar = document.getElementById('showShellResult');
        if (scrollbar) {
            const wrap = scrollbar.parentNode;
            //往上移了就不滚动了
            let distanceToBottom = wrap ? wrap.scrollHeight - wrap.scrollTop - wrap.clientHeight : 0;
            console.log('高度' , force , distanceToBottom)
            if(!force && distanceToBottom > 30){
                // return
            }
            if (wrap) {
                wrap.scrollTop = wrap.scrollHeight;
                console.log(wrap.scrollTop)
            }
        }


        // let obj = document.getElementById('showShellResult')
        // if (obj) {
        //     obj.scrollTop = obj.scrollHeight + 200
        // }
    } , 200)
}



function ShellOutStart(data , callBack){
    base.BasePost('/api/shellOut', data, callBack)
}

function ShellOutSetSeeId(data , callBack){
    base.BasePost('/api/shellOutSetSeeId', data, callBack)
}

function ShellOutEdit(data , callBack){
    base.BasePost('/api/shellOutEdit',data, callBack)
}

export default {
    calculateShellDivHeight,
    ShellDivToBottom,
    ShellOutStart,
    ShellOutSetSeeId,
    ShellOutEdit,
}