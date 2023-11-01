import base from './base'
import mod from './module'

//php的序列化值反序列化
function PhpUnserialize(serializeStr , callBack){
  base.BasePost('/api/PhpUnserialize', {
    SerializeStr : serializeStr,
  }, function (response) {
    callBack(response)
  })
}
export default {
  PhpUnserialize,
}
