(function() {
    setTimeout(function() {
        // 删除已存在的浮动块（如果存在）
        const existingFloater = document.getElementById('cookie-floater');
        if (existingFloater) {
            existingFloater.remove();
        }
        let cookieList = document.cookie.split(';')
        let showContent = ''
        let config = {config}
        //获取配置方法
        const getShowValue = function (item , cookieList){
            let findValue = '' //查找的值
            let isFind = false //是否找到了
            cookieList.forEach(cookie => {
                if(isFind){
                    return
                }
                let [name, value] = cookie.split('=');
                name = name.trimStart()
                //进行格式化
                if(item.format_list && item.format_list.length > 0){
                    for(let i = 0; i < item.format_list.length;i++){
                        if(item.format_list[i] === "url_decode"){
                            try {
                                value = decodeURIComponent(value)
                            } catch (e) {
                            }
                        }
                    }
                }
                if(item.find_type === 'cookie') { //直接cookie值匹配
                    if(name === item.find_key){
                        isFind = true
                        findValue = value
                    }
                }else if(item.find_type === 'any'){ //任意值
                    //查找
                    if(item.regex_find_key !== ''){
                        let regexStr = new RegExp(item.regex_find_key)
                        let match = value.match(regexStr);
                        if (match && match[1]) {
                            findValue = match[1]
                            isFind = true
                        }
                    }else{
                        console.log('不支持的查找方式')
                    }
                }
            });
            return findValue
        }
        // 遍历配置项获取Cookie值
        config.forEach(item => {
            showContent += '<span>'+item.label+': <strong>'+getShowValue(item , cookieList)+'</strong></span>&nbsp;'

        });
        // 创建浮动块
        const floater = document.createElement('div');
        floater.id = 'cookie-floater';
        floater.style.position = 'fixed';
        floater.style.bottom = '0px';
        floater.style.padding = '3px';
        floater.style.backgroundColor = '#D0EBFF';
        // floater.style.border = '1px solid #D0EBFF';
        floater.style.borderRadius = '4px';
        floater.style.boxShadow = '0 2px 10px rgba(0,0,0,0.1)';
        floater.style.zIndex = '9999';
        floater.style.fontSize = '12px';
        floater.style.opacity  = "0.9"
        // floater.style.transition = "all 0.5s ease-in-out";

        // 从 localStorage 读取位置状态（默认右侧）
        // const savedPosition = localStorage.getItem('cookieFloaterPosition') || 'right';
        // if (savedPosition === 'left') {
        //     floater.style.left = '20px';
        //     floater.style.right = 'auto';
        // } else {
        //     floater.style.right = '20px';
        //     floater.style.left = 'auto';
        // }
        // ---------- 随机位置工具 ----------
        // 产生 [min, max] 随机整数
        let rand = (min, max) => Math.floor(Math.random() * (max - min + 1)) + min;

        /**
         * 把 el 随机放到可视区域内，不改变它原来的宽高
         * @param {HTMLElement} el
         */
        function randomPos(el) {
            const pad = 20;                                      // 留边距
            const maxX = window.innerWidth - el.offsetWidth - pad;
            const maxY = window.innerHeight - el.offsetHeight - pad;

            const x = rand(pad, Math.max(pad, maxX));
            const y = rand(pad, Math.max(pad, maxY));

            /* 关键：只写 left/top，right/bottom 必须 auto */
            el.style.position = 'fixed';
            el.style.left = x + 'px';
            el.style.top  = y + 'px';
            el.style.right = 'auto';
            el.style.bottom = 'auto';

            /* 记忆坐标，刷新后还原 */
            localStorage.setItem('cookieFloaterPos', JSON.stringify({ x, y }));
        }
        const saved = localStorage.getItem('cookieFloaterPos');
        if (saved) {
            const { x, y } = JSON.parse(saved);
            floater.style.position = 'fixed';
            floater.style.left = x + 'px';
            floater.style.top = y + 'px';
            floater.style.right = 'auto';
            floater.style.bottom = 'auto';
        } else {
            randomPos(floater);                              // 第一次随机
        }

        // 添加内容
        let html = '<div style="display: flex; justify-content: space-between; align-items: center;">'
        html += '<button id="open-new-tab" style="background: none; border: none; cursor: pointer; font-size: 12px; color: #666; padding: 2px 5px;">新页卡</button>';
        html += '<div>' + showContent + '</div>';
        html += '<div style="display: flex; flex-direction: column; gap: 5px;">';
        html += '<button id="move-floater" style="background: none; border: none; cursor: pointer; font-size: 12px; color: #666; padding: 2px 5px;">滚开</button>';
        //html += '<button id="close-floater" style="background: none; border: none; cursor: pointer; font-size: 12px; color: #666; padding: 2px 5px;">关闭</button>';
        html += '</div>';
        html += '</div>';
        floater.innerHTML = html;

        // 添加移动按钮事件
        floater.querySelector('#move-floater').addEventListener('click', () => {
            randomPos(floater);
        });

        floater.querySelector('#open-new-tab').addEventListener('click', () => {
            var curUrl = window.location.href;
            window.open(curUrl, '_blank');
        });

        // // 添加关闭按钮事件
        // floater.querySelector('#close-floater').addEventListener('click', () => {
        //     floater.remove();
        // });

        // 插入到页面
        document.body.appendChild(floater);
    }, 1000);
})();