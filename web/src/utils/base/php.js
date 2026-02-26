import base from '../base'

//php的序列化值反序列化
function PhpUnserialize(serializeStr, callBack) {
  base.BasePost(
    '/api/PhpUnserialize',
    {
      SerializeStr: serializeStr,
    },
    function (response) {
      callBack(response)
    }
  )
}

//php的序列化值反序列化
function PhpUnserialize2(serializeStr, callBack) {
  base.BasePost(
      '/api/PhpUnserialize2',
      {
        SerializeStr: serializeStr,
      },
      function (response) {
        callBack(response)
      }
  )
}
export default {
  PhpUnserialize,
  PhpUnserialize2,
}
