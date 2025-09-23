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
        /* ========== 统一漂亮样式 ========== */
        Object.assign(floater.style, {
            position: 'fixed',
            zIndex: '9999',
            padding: '10px 14px',
            fontSize: '13px',
            fontFamily: 'system-ui, Segoe UI, Roboto, sans-serif',
            color: '#0b2253',
            background: 'rgba(255,255,255,0.65)',
            backdropFilter: 'blur(12px) saturate(180%)',
            WebkitBackdropFilter: 'blur(12px) saturate(180%)',
            border: '1px solid rgba(255,255,255,0.35)',
            borderRadius: '10px',
            boxShadow: '0 8px 32px rgba(31, 38, 135, 0.25)',
            transition: 'all .35s cubic-bezier(.4,0,.2,1)',
            cursor: 'default'
        });

        /* 按钮统一风格 */
        const btnStyle = `
            background: rgba(30,111,255,0.12);
            border: 1px solid rgba(30,111,255,0.22);
            border-radius: 6px;
            color: #1e6fff;
            padding: 4px 8px;
            margin: 0 2px;
            font-size: 12px;
            cursor: pointer;
            transition: background .25s;
        `;

        /* ========== 四选一随机定位 ========== */
        const POSITIONS = [
            { name: 'bottom-left',  apply(el) { el.style.left = '20px';  el.style.bottom = '20px'; el.style.top = el.style.right = 'auto'; } },
            { name: 'bottom-right', apply(el) { el.style.right = '20px'; el.style.bottom = '20px'; el.style.top = el.style.left = 'auto'; } },
            { name: 'top-left',     apply(el) { el.style.left = '20px';  el.style.top = '20px';    el.style.bottom = el.style.right = 'auto'; } },
            { name: 'top-right',    apply(el) { el.style.right = '20px'; el.style.top = '20px';    el.style.bottom = el.style.left = 'auto'; } }
        ];

        /* 读取当前位置名称 */
        function getCurrentPosName(el) {
            const l = el.style.left, r = el.style.right, t = el.style.top, b = el.style.bottom;
            if (l === '20px' && b === '20px') return 'bottom-left';
            if (r === '20px' && b === '20px') return 'bottom-right';
            if (l === '20px' && t === '20px') return 'top-left';
            if (r === '20px' && t === '20px') return 'top-right';
            return null;
        }

        /* 随机换到“另外三个位置”中的任意一个 */
        function randomPos(el) {
            const current = getCurrentPosName(el);
            const candidates = POSITIONS.filter(p => p.name !== current);
            const picked = candidates[Math.floor(Math.random() * candidates.length)];
            picked.apply(el);
            /* 记忆 */
            localStorage.setItem('cookieFloaterPos_v1', JSON.stringify({ pos: picked.name }));
        }

        /* ========== 初始化时还原上次位置 ========== */
        const saved = localStorage.getItem('cookieFloaterPos_v1');
        if (saved) {
            const { pos } = JSON.parse(saved);
            const p = POSITIONS.find(i => i.name === pos);
            if (p) p.apply(floater);
        } else {
            randomPos(floater);          // 第一次随机
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

        floater.querySelectorAll('button').forEach(b => b.setAttribute('style', btnStyle));

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