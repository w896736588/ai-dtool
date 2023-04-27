import Vue from "vue";

function ajaxDefault(params , callBack){
  let apiHost = Vue.prototype.$helperConfig.getApiHost()
  Vue.axios.post(apiHost + '/api/shell/exec', params).then(function (response) {
    callBack(response)
  });
}

export default {
  ajaxDefault,
}
