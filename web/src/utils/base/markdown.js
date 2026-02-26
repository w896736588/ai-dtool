import base from '../base'

function MarkdownList(markdownType , callBack){
    base.BasePost('/api/MarkdownList', {
        markdown_type : markdownType,
    } , callBack)
}
function MarkdownSort(markdown_ids,callBack){
    base.BasePost('/api/MarkdownSort', {
        markdown_ids : markdown_ids,
    } , callBack)
}

function MarkdownHistoryList(markdownId , callBack){
    base.BasePost('/api/MarkdownHistoryList', {
        id : markdownId,
    } , callBack)
}


function MarkdownAdd(id ,markdownType, name ,content , callBack){
    base.BasePost('/api/MarkdownAdd', {
        id : id,
        markdown_type : markdownType,
        name : name,
        content : content,
    } , callBack)
}

function MarkdownDel(id , callBack){
    base.BasePost('/api/MarkdownDel', {id : id} , callBack)
}

function MarkdownHistoryDel(id , callBack){
    base.BasePost('/api/MarkdownHistoryDel', {id : id} , callBack)
}

export default {
    MarkdownList,
    MarkdownAdd,
    MarkdownDel,
    MarkdownHistoryList,
    MarkdownHistoryDel,
    MarkdownSort,
}