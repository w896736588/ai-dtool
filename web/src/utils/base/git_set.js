import base from "@/utils/base";

function GitList(callBack){
    base.BasePost('/api/Set/GitList', {} , callBack)
}
function GitAdd(data , callBack){
    base.BasePost('/api/Set/GitAdd', data, callBack)
}
function GitDelete(data , callBack){
    base.BasePost('/api/Set/GitDelete', data, callBack)
}

function GitGroupList(callBack){
    base.BasePost('/api/Set/GitGroupList', {} , callBack)
}
function GitGroupAdd(data , callBack){
    base.BasePost('/api/Set/GitGroupAdd', data, callBack)
}
function GitGroupDelete(data , callBack){
    base.BasePost('/api/Set/GitGroupDelete', data, callBack)
}

function GitQuickList(data , callBack){
    base.BasePost('/api/Set/GitQuickList', data, callBack)
}


function GitlabTokenList(callBack){
    base.BasePost('/api/Set/GitLabTokenList', {} , callBack)
}
function GitlabTokenAdd(data , callBack){
    base.BasePost('/api/Set/GitLabTokenCreate', data, callBack)
}
function GitlabTokenDelete(data , callBack){
    base.BasePost('/api/Set/GitLabTokenDelete', data, callBack)
}

function GlobalList(callBack){
    base.BasePost('/api/Set/GlobalList', {} , callBack)
}
function GlobalAdd(data , callBack){
    base.BasePost('/api/Set/GlobalCreate', data, callBack)
}
function GlobalDelete(data , callBack){
    base.BasePost('/api/Set/GlobalDelete', data, callBack)
}

function MemoryConfigGet(callBack){
    base.BasePost('/api/Set/MemoryConfigGet', {}, callBack)
}

function MemoryConfigSave(data , callBack){
    base.BasePost('/api/Set/MemoryConfigSave', data, callBack)
}

function RuntimeConfigSave(data , callBack){
    base.BasePost('/api/Set/RuntimeConfigSave', data, callBack)
}

function RuntimeDatabaseGitSync(data , callBack){
    base.BasePost('/api/Set/RuntimeDatabaseGitSync', data, callBack)
}

function RuntimeConfigItemSave(data , callBack){
    base.BasePost('/api/Set/RuntimeConfigItemSave', data, callBack)
}

function CronConfigGet(callBack){
    base.BasePost('/api/Set/CronConfigGet', {}, callBack)
}

function CronConfigSave(data , callBack){
    base.BasePost('/api/Set/CronConfigSave', data, callBack)
}

export default {
    GitList,
    GitAdd,
    GitDelete,
    GitGroupList,
    GitGroupAdd,
    GitGroupDelete,
    GitQuickList,
    GitlabTokenList,
    GitlabTokenAdd,
    GitlabTokenDelete,
    GlobalList,
    GlobalAdd,
    GlobalDelete,
    MemoryConfigGet,
    MemoryConfigSave,
    RuntimeConfigSave,
    RuntimeDatabaseGitSync,
    RuntimeConfigItemSave,
    CronConfigGet,
    CronConfigSave,
}
