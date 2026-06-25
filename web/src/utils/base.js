//登录
import store from './base/store'
import module from './module'
import {globals} from '@/main'

const DEV_PORT = '17170'
const SAFE_TOKEN_KEY = 'safe_token'
let runtimeSsePort = ''

// 获取服务端注入的配置（Go模板渲染时注入）
function GetServerConfig() {
    return window.__SERVER_CONFIG__ || null
}

// 获取 safe token
function GetSafeToken() {
    return store.getStore(SAFE_TOKEN_KEY) || ''
}

// 设置 safe token
function SetSafeToken(token) {
    if (token) {
        store.setStore(SAFE_TOKEN_KEY, token)
    }
}

function SetSsePort(port) {
    runtimeSsePort = port ? String(port).trim() : ''
}

// 清除 safe token
function ClearSafeToken() {
    store.removeStore(SAFE_TOKEN_KEY)
}

// Safe 登录接口（使用配置文件密码）
function BaseLogin(password, okFunc) {
    BasePost(
        '/api/BaseLogin',
        {
            password: password,
        },
        function (response) {
            if (response.ErrCode === 0 && response.Data && response.Data.token) {
                SetSafeToken(response.Data.token)
            }
            okFunc(response)
        }
    )
}

// 检查登录状态
function BaseLoginStatus(okFunc) {
    BasePost(
        '/api/BaseLoginStatus',
        {},
        function (response) {
            okFunc(response)
        }
    )
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

// 处理响应，检查续期 token 和登录失效
function handleResponse(response) {
    // 检查是否有续期 token
    const renewedToken = response.headers && response.headers['x-renewed-token']
    if (renewedToken) {
        SetSafeToken(renewedToken)
    }
    return response
}

// 处理错误，检查登录失效错误码
function handleError(error, callBack) {
    const cb = typeof callBack === 'function' ? callBack : null
    if (error.response && error.response.data) {
        const errCode = error.response.data.ErrCode
        const errMsg = error.response.data.ErrMsg || ''
        // 40101: 未登录, 40102: 过期, 40103: 密码版本不匹配, 40104: token非法
        if (errCode === 40101 || errCode === 40102 || errCode === 40103 || errCode === 40104) {
            ClearSafeToken()
            // 触发全局登录失效事件，带上错误消息用于弹窗显示
            if (globals && globals.$eventBus) {
                globals.$eventBus.emit('safe_auth_required', { message: errMsg })
            }
            // 登录失效时调用回调，但将错误消息置空，避免显示错误通知
            // 40103（密码版本不匹配）通常是修改密码后触发，直接弹窗不显示错误
            if (cb) cb({...error.response.data, ErrMsg: '', __loginRequired: true})
            return
        }
        if (cb) cb(error.response.data)
        return
    }
    if (cb) cb({ErrCode: -1, ErrMsg: error.message || '请求失败'})
}

//POST请求
function BasePost(uri, params, callBack) {
    globals.$axios.post(GetApiHost() + uri, params , {
        headers: {
            'Content-Type': 'application/json',
            'Token' : GetSafeToken(),
        }
    }).then(function (response) {
        callBack(handleResponse(response))
    }).catch(function (error) {
        handleError(error, callBack)
    })
}

//POST请求 with multipart/form-data
function BasePostForm(uri, params, callBack) {
    globals.$axios.post(GetApiHost() + uri, params , {
        headers: {
            'Content-Type': 'multipart/form-data',
            'Token' : GetSafeToken(),
        }
    }).then(function (response) {
        callBack(handleResponse(response))
    }).catch(function (error) {
        handleError(error, callBack)
    })
}

// 判断是否为开发环境
function isDev() {
    return process.env.NODE_ENV === 'development'
}

// 获取基础 API 地址
// 开发环境：固定使用 localhost:17170
// 生产环境：使用相对路径（同域）
function GetApiHost() {
    if (isDev()) {
        const config = GetServerConfig()
        const port = (config && config.port) ? config.port : DEV_PORT
        return 'http://localhost:' + port
    }
    return ''  // 生产环境返回空字符串，使用相对路径
}

// 获取完整的 API 基地址（包含协议、主机和端口），用于需要拼接绝对路径的场景。
// 生产环境直接取地址栏的 origin，开发环境与 GetApiHost 一致。
function GetAbsoluteApiHost() {
    if (isDev()) {
        return GetApiHost()
    }
    return window.location.origin
}

// 获取 SSE API 地址
function GetSseApiHost(port) {
    const ssePort = port || runtimeSsePort
    if (!ssePort) {
        return ''
    }
    if (isDev()) {
        return 'http://localhost:' + ssePort
    }
    // 生产环境：用 location.hostname 拼接 SSE 端口
    return window.location.protocol + '//' + window.location.hostname + ':' + ssePort
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
    const rand = Math.floor(Math.random() * 100000000); // 8 位随机数字
    return prefix + `_${String(rand).padStart(8, '0')}`;
}

function GenerateSseClientId(prefix) {
    const rand = Math.floor(Math.random() * 100000000); // 8 位随机数字
    return prefix + `_${String(rand).padStart(8, '0')}`;
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
    BaseLoginStatus,
    BasePost,
    BasePostForm,
    GetApiHost,
    GetAbsoluteApiHost,
    GetSseApiHost,
    GetServerConfig,
    GetSafeToken,
    SetSafeToken,
    ClearSafeToken,
    SetSsePort,
    Globals,
    GetDivHeight,
    GetDivHeight2,
    IsBase64,
    GenerateId,
    GenerateSseClientId,
    Debounce,
    DisableSaveShortcut,
    UploadFile,
    FormatEnterToMarkdown,
}
