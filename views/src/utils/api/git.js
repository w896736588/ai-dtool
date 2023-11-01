import base from './base'
import module from './module'
//查询分支
function GitCurrentBranch(chooseCodeUnikey , callBack){
  let codeConfig = GitGetCodeConfigByUnikey(chooseCodeUnikey)
  base.BasePost('/api/GitQueryCurrentBranch', {
    CodePath : codeConfig.CodePath,
    ShellName : codeConfig.SshName,
  }, function (response) {
    callBack(response)
  })
}

//拉取最新代码
function GitPullBranchOrigin(chooseCodeUnikey , callBack){
  let codeConfig = GitGetCodeConfigByUnikey(chooseCodeUnikey)
  base.BasePost('/api/GitPullBranchOrigin', {
    CodePath : codeConfig.CodePath,
    ShellName : codeConfig.SshName,
  }, function (response) {
    callBack(response)
  })
}

function GitChangeBranch(chooseCodeUnikey ,branchName, callBack){
  let codeConfig = GitGetCodeConfigByUnikey(chooseCodeUnikey)
  base.BasePost('/api/GitChangeBranch', {
    CodePath : codeConfig.CodePath,
    ShellName : codeConfig.SshName,
    BranchName : branchName,
  }, function (response) {
    callBack(response)
  })
}

function GitQueryStatus(chooseCodeUnikey, callBack){
  let codeConfig = GitGetCodeConfigByUnikey(chooseCodeUnikey)
  base.BasePost('/api/GitQueryStatus', {
    CodePath : codeConfig.CodePath,
    ShellName : codeConfig.SshName,
  }, function (response) {
    callBack(response)
  })
}

function GitGetCodeConfigByUnikey(chooseCodeUnikey){
  let codeConfigList = module.GetCodeConfigList()
  for(let i in codeConfigList){
    if(codeConfigList[i].Unikey === chooseCodeUnikey){
      return codeConfigList[i];
    }
  }
}

export default {
  GitCurrentBranch,
  GitPullBranchOrigin,
  GitChangeBranch,
  GitQueryStatus,
  GitGetCodeConfigByUnikey,
}
