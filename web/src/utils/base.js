//登录
import store from './base/store'
import module from './module'
import {globals} from '@/main'
import apiHostRouter from './api_host_router.cjs'

var ports = ["17170"]
const ssePort = '17170'

//登录拿到 unikey
function BaseLogin(userName, password, okFunc) {
    BasePost(
        '/api/BaseLogin',
        {
            UserName: userName,
            Password: password,
        },
        function (response) {
            okFunc(response)
        }
    )
}

function Ports() {
    BasePost('/api/ports', {}, function (response) {
        if (response.Data && response.Data.ports) {
            ports = response.Data.ports
        }
    })
}


function SetIsInit(pageName) {
    Globals().$pageInit[pageName] = true
}

function GetIsInit(pageName) {
    if (Globals().$pageInit[pageName]) {
        return Globals().$pageInit[pageName]
    } else {
        return false
    }
}

function Globals() {
    return globals
}

//POST请求
function BasePost(uri, params, callBack) {
    globals.$axios.post(GetApiHost() + uri, params , {
        headers: {
            'Content-Type': 'application/json', // 常用示例
            'Token' : store.getStore('token'),
        }
    }).then(function (response) {
        callBack(response)
    })
}

//POST请求 with multipart/form-data
function BasePostForm(uri, params, callBack) {
    globals.$axios.post(GetApiHost() + uri, params , {
        headers: {
            'Content-Type': 'multipart/form-data',
            'Token' : store.getStore('token'),
        }
    }).then(function (response) {
        callBack(response)
    })
}

//拿到接口地址
function GetApiHost() {
    let port = GetApiPort()
    return 'http://localhost:' + port
}

function GetSseApiHost() {
    let port = GetSsePort()
    return 'http://localhost:' + port
}

function GetApiPort() {
    return apiHostRouter.getApiPort(ports, Math.random, ssePort)
}

function GetSsePort() {
    return apiHostRouter.getSsePort(ports, Math.random, ssePort)
}

function GetRandPort() {
    return GetApiPort()
}

//上面是mainCard 这个返回mainCard距离底部还剩余的高度px
function GetDivHeight() {
    let mainCard = document.getElementById('mainCard')
    if (!mainCard) {
        return
    }
    let rect = mainCard.getBoundingClientRect();
    let viewportHeight = window.innerHeight || document.documentElement.clientHeight;
    return viewportHeight - rect.bottom - 25;
}

function GetDivHeight2() {
    let mainCard = document.getElementById('mainCard')
    if (!mainCard) {
        return
    }
    let rect = mainCard.getBoundingClientRect();
    let viewportHeight = window.innerHeight || document.documentElement.clientHeight;
    return viewportHeight - rect.top - 25;
}

function IsBase64(str) {
    // 排除纯数字字符串
    if (/^\d+$/.test(str)) {
        return false;
    }
    // 正则表达式匹配 Base64 编码
    const base64Regex = /^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{4}|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)$/;
    if (!base64Regex.test(str)) {
        return false;
    }
    // 尝试解码
    try {
        const decoded = atob(str);
        // 重新编码，检查是否与原始字符串一致
        const reencoded = btoa(decoded);
        return reencoded === str
    } catch (e) {
        return false;
    }
}

function GenerateId(prefix) {
    const ts = Date.now();                 // 毫秒时间戳
    const rand = Math.floor(Math.random() * 100) + 100; // 3 位随机数
    return prefix + `_${rand}`;
}

//防抖函数
function Debounce(fn, delay = 500) {
    let t = null;
    return (...a) => {
        clearTimeout(t);
        t = setTimeout(() => fn(...a), delay);
    };
}

//全局禁用ctrl+s
function DisableSaveShortcut() {
    document.addEventListener('keydown', function (e) {
        // 检测 Ctrl+S 或 Cmd+S
        if ((e.ctrlKey || e.metaKey) &&
            (e.key === 's' || e.keyCode === 83 || e.which === 83)) {
            e.preventDefault();
            e.stopPropagation();
            console.log('已禁用ctrl+s')
            return false;
        }
    });
}


//POST请求
function UploadFile(file, callBack) {
    const form = new FormData()
    form.append('file', file)
    form.append('Unikey', store.getStore('Unikey'))
    form.append('token', store.getStore('Unikey'))
    globals.$axios.post(GetApiHost() + '/api/Upload', form, {
        headers: {'Content-Type': undefined}   // 让 axios 自动带 multipart
    }).then(function (response) {
        callBack(response)
    })
}

function FormatEnterToMarkdown(data) {
    if (!data) {
        return ''
    }
    return data.replace(/\r\n|\n|\r/g, '<br>')
        .replace(/\s+/g, ' ') // 可选：合并多个空格
        .trim()
}

export default {
    BaseLogin,
    BasePost,
    BasePostForm,
    GetApiHost,
    GetSseApiHost,
    Globals,
    GetDivHeight,
    GetDivHeight2,
    IsBase64,
    GenerateId,
    GetRandPort,
    GetApiPort,
    GetSsePort,
    Debounce,
    Ports,
    DisableSaveShortcut,
    UploadFile,
    FormatEnterToMarkdown,
}
