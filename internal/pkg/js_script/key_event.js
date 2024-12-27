function setupF12Listener() {
    document.addEventListener('keydown', (event) => {
        console.log(event);
        if (event.key === 'F12') {
            console.log('F12 key pressed');
        }
    });
}
setupF12Listener();

function addTipMsg(msg){
    var messageBox = document.createElement('div');
    messageBox.id = 'messageBox';
    messageBox.textContent = '` + tip + `';
    messageBox.style.position = 'fixed';
    messageBox.style.top = '50%';
    messageBox.style.left = '50%';
    messageBox.style.transform = 'translate(-50%, -50%)';
    messageBox.style.color = 'white';
    messageBox.style.backgroundColor = 'black';
    messageBox.style.padding = '20px';
    messageBox.style.borderRadius = '10px';
    messageBox.style.boxShadow = '0 0 10px rgba(0, 0, 0, 0.5)';
    messageBox.style.zIndex = 2000;
    messageBox.style.display = 'block'; // 初始状态隐藏
    document.body.appendChild(messageBox);
    setTimeout(function() {
        messageBox.style.display = 'none'; // 隐藏消息框
    }, 2000); // 5000毫秒等于5秒
}
console.log('js加载完成')