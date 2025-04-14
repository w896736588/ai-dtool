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
        // 遍历配置项获取Cookie值
        config.forEach(item => {
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

            // 创建每行显示
            showContent += '<div>'+item.label+': <strong>'+findValue+'</strong></div>'

        });
        // 创建浮动块
        const floater = document.createElement('div');
        floater.id = 'cookie-floater';
        floater.style.position = 'fixed';
        floater.style.bottom = '20px';
        floater.style.right = '20px';
        floater.style.padding = '10px';
        floater.style.backgroundColor = '#f0f0f0';
        floater.style.border = '1px solid #ccc';
        floater.style.borderRadius = '4px';
        floater.style.boxShadow = '0 2px 10px rgba(0,0,0,0.1)';
        floater.style.zIndex = '9999';

        // 添加内容
        let html = '<div style=" justify-content: space-between; align-items: center;">'
        html += showContent;
        html += '<button ';
        html += ' id="close-floater" '
        html += ' style="background: none;border: none;margin-left:47%;cursor: pointer;font-size: 14px;color: #666;">关闭</button> '
        html += ' </div> '
        floater.innerHTML = html

        // 添加关闭按钮事件
        floater.querySelector('#close-floater').addEventListener('click', () => {
            floater.remove();
        });

        // 插入到页面
        document.body.appendChild(floater);
    }, 1000);
})();