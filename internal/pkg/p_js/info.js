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
            showContent += '<div>'+item.label+': <strong>'+getShowValue(item , cookieList)+'</strong></div>'

        });
        // 创建浮动块
        const floater = document.createElement('div');
        floater.id = 'cookie-floater';
        floater.style.position = 'fixed';
        floater.style.bottom = '20px';
        floater.style.padding = '10px';
        floater.style.backgroundColor = '#f0f0f0';
        floater.style.border = '1px solid #ccc';
        floater.style.borderRadius = '4px';
        floater.style.boxShadow = '0 2px 10px rgba(0,0,0,0.1)';
        floater.style.zIndex = '9999';
        floater.style.transition = 'all 0.3s ease'; // 添加过渡效果

        // 从 localStorage 读取位置状态（默认右侧）
        const savedPosition = localStorage.getItem('cookieFloaterPosition') || 'right';
        if (savedPosition === 'left') {
            floater.style.left = '20px';
            floater.style.right = 'auto';
        } else {
            floater.style.right = '20px';
            floater.style.left = 'auto';
        }

        // 添加内容
        let html = '<div style="display: flex; justify-content: space-between; align-items: center;">'
        html += '<div>' + showContent + '</div>';
        html += '<div style="display: flex; flex-direction: column; gap: 5px;">';
        html += '<button id="move-floater" style="background: none; border: none; cursor: pointer; font-size: 12px; color: #666; padding: 2px 5px;">滚开</button>';
        html += '<button id="close-floater" style="background: none; border: none; cursor: pointer; font-size: 12px; color: #666; padding: 2px 5px;">关闭</button>';
        html += '</div>';
        html += '</div>';
        floater.innerHTML = html;

        // 添加移动按钮事件
        let isRight = true;
        floater.querySelector('#move-floater').addEventListener('click', () => {
            const currentPosition = floater.style.left === '20px' ? 'left' : 'right';
            const newPosition = currentPosition === 'right' ? 'left' : 'right';

            if (newPosition === 'left') {
                floater.style.left = '20px';
                floater.style.right = 'auto';
            } else {
                floater.style.right = '20px';
                floater.style.left = 'auto';
            }

            // 存储新位置到 localStorage
            localStorage.setItem('cookieFloaterPosition', newPosition);
        });

        // 添加关闭按钮事件
        floater.querySelector('#close-floater').addEventListener('click', () => {
            floater.remove();
        });

        // 插入到页面
        document.body.appendChild(floater);
    }, 1000);
})();