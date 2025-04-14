(function() {
    setTimeout(function() {
        let existTip = document.getElementById('playwrightTipId');
        if (existTip) {
            existTip.remove();
        }
        var messageBox = document.createElement('div');
        messageBox.id = 'playwrightTipId';
        messageBox.textContent = '{tip}';
        messageBox.style.position = 'fixed';
        messageBox.style.top = '70%';
        messageBox.style.left = '50%';
        messageBox.style.transform = 'translate(-50%, -50%)';
        messageBox.style.color = 'white';
        messageBox.style.backgroundColor = 'black';
        messageBox.style.padding = '15px';
        messageBox.style.borderRadius = '10px';
        messageBox.style.boxShadow = '0 0 10px rgba(0, 0, 0, 0.5)';
        messageBox.style.zIndex = '2000';
        messageBox.style.display = 'block'; // 初始状态隐藏
        document.body.appendChild(messageBox);
        setTimeout(function() {
            let existTip = document.getElementById('playwrightTipId');
            if (existTip) {
                existTip.remove();
            }
        }, 2000);
    }, 100);
})();