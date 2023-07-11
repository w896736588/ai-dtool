function getCurrentDateTime(){
  // 创建一个新的Date对象，代表当前日期和时间
  let currentDate = new Date();
  // 使用Date对象的方法获取小时、分钟、秒
  let hours = currentDate.getHours();
  let minutes = currentDate.getMinutes();
  let seconds = currentDate.getSeconds();
  // 将小时、分钟、秒拼接成字符串，并添加0前缀（如果需要）
  return hours.toString().padStart(2, '0') + ":" +
    minutes.toString().padStart(2, '0') + ":" +
    seconds.toString().padStart(2, '0');
}
export default {
  getCurrentDateTime,
}
