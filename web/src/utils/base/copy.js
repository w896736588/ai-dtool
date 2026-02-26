import notify from "./notify"
import t from "./type"
let copyList = []

function SetCopyContent(content){
    if(t.IsObjectOrArray(content)){
        content = JSON.stringify(content)
    }
    for (let i = 0; i < copyList.length; i++) {
        if (copyList[i] === content){
            return i
        }
    }
    copyList.push(content)
    console.log('当前复制内容为' , content, copyList.length)
    return copyList.length - 1
}
function handleCopy(copyIndex) {
    console.log(copyIndex , copyList)
    let copyText = copyList[parseInt(copyIndex)]
    // 创建一个临时的textarea用于辅助复制操作
    const textarea = document.createElement('textarea');
    textarea.style.position = 'absolute';
    textarea.style.left = '-9999px';
    textarea.value = copyText;
    document.body.appendChild(textarea);

    // 将文本选中并执行复制操作
    textarea.select();
    document.execCommand('copy'); // 执行浏览器内置的复制命令

    // 清理临时创建的textarea
    document.body.removeChild(textarea);
    notify.success('复制成功')
}

export default {
    handleCopy,
    SetCopyContent,
}