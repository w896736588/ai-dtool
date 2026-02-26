import base from "@/utils/base";

function Ai(data , callBack){
    base.BasePost('/api/AiRun', data , callBack)
}

export default {
    Ai,
}