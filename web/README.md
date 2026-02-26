npm config set registry https://registry.npmmirror.com/
开发计划：
1.git，supervisor支持socket断开后，执行任何命令时自动建立新的链接（链接断开后要删除本地缓存的句柄） 
20241210 已完成
2.自动化链接，支持页面显示当前已打开的浏览器并进行控制关闭
20241210 已完成
3.增加总览信息，可以查看目前已连接的所有ssh，redis，mysql状态，可以选择重连或断开
4.合并工具，包括二维码，sql生成model，时间转换器
5.新增工具：
    json解析：支持解析一次，解析无限次（递归里面所有的字符串进行json解析）
    urldecode
    unidecode