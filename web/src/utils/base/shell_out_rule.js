import base from "@/utils/base";

function ShellOutRuleSetList(data, callBack) {
    base.BasePost('/api/ShellOutRuleSetList', data, callBack)
}

function ShellOutRuleSetInfo(data, callBack) {
    base.BasePost('/api/ShellOutRuleSetInfo', data, callBack)
}

function ShellOutRuleSetSave(data, callBack) {
    base.BasePost('/api/ShellOutRuleSetSave', data, callBack)
}

function ShellOutRuleSetDelete(data, callBack) {
    base.BasePost('/api/ShellOutRuleSetDelete', data, callBack)
}

function ShellOutRuleImportLegacy(data, callBack) {
    base.BasePost('/api/ShellOutRuleImportLegacy', data, callBack)
}

export default {
    ShellOutRuleSetList,
    ShellOutRuleSetInfo,
    ShellOutRuleSetSave,
    ShellOutRuleSetDelete,
    ShellOutRuleImportLegacy,
}
