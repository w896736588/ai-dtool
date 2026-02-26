function setStore(key, value) {
  localStorage.setItem(key, value)
}

function getStore(key) {
  return localStorage.getItem(key)
}

function GetStoreIdInt(key){
  let cacheData = localStorage.getItem(key)
  if(cacheData === null || cacheData === undefined){
    return 0
  }
  return parseInt(cacheData)
}

function removeStore(key){
  localStorage.removeItem(key)
}

export default {
  setStore,
  getStore,
  removeStore,
  GetStoreIdInt,
}
