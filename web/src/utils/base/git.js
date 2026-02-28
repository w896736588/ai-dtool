import base from '../base'
import module from '../module'
import store from './store'


function GitConfigList(config,callBack) {
    base.BasePost('/api/GitConfigList', config, callBack)
}

//查询分支
function GitCurrentBranch(gitConfig, callBack) {
    base.BasePost('/api/GitQueryCurrentBranch', gitConfig, callBack)
}

//拉取最新代码
function GitPullBranchOrigin(gitConfig, callBack) {
    base.BasePost('/api/GitPullBranchOrigin', gitConfig, callBack)
}

function GitChangeBranch(gitConfig, branchName, callBack) {
    gitConfig.BranchName = branchName
    base.BasePost('/api/GitChangeBranch', gitConfig, callBack)
}

function GitChangeBranchRemote(gitConfig, branchName, callBack) {
    gitConfig.BranchName = branchName
    base.BasePost('/api/GitChangeBranchRemote', gitConfig, callBack)
}

function SetSafe(gitConfig , callBack){
    base.BasePost('/api/GitSetSafeLog', gitConfig, callBack)
}

function GitSaveCredentials(gitConfig , callBack){
    base.BasePost('/api/GitSaveCredentials', gitConfig, callBack)
}

function GitQueryStatus(gitConfig, callBack) {
    base.BasePost('/api/GitQueryStatus',gitConfig, callBack)
}

function GitLocalSetLastGroupId(groupId){
    store.setStore('last_group_id' , groupId)
}

function GitLocalSetLastGitId(gitId){
    store.setStore('last_git_id' , gitId)
}

function GitLocalGetLastGroupId(){
    let lastGroupId = store.getStore('last_group_id')
    if(lastGroupId === '' || lastGroupId === null || lastGroupId === undefined){
        return 0
    }
    return lastGroupId
}
function GitLocalGetLastGitId(){
    let lastGitId = store.getStore('last_git_id')
    if(lastGitId === '' || lastGitId === null || lastGitId === undefined){
        return 0
    }
    return lastGitId
}

//查询分支
function GitCommitLog(gitConfig , callBack) {
    base.BasePost('/api/GitCommitLog', gitConfig, callBack)
}

function GitGroupBranchList(data, callBack) {
    base.BasePost('/api/GitGroupBranchList', data, callBack)
}

export default {
    GitCurrentBranch,
    GitPullBranchOrigin,
    GitChangeBranch,
    GitQueryStatus,
    GitCommitLog,
    GitConfigList,
    GitLocalSetLastGroupId,
    GitLocalGetLastGroupId,
    GitLocalSetLastGitId,
    GitLocalGetLastGitId,
    GitChangeBranchRemote,
    SetSafe,
    GitSaveCredentials,
    GitGroupBranchList,
}
