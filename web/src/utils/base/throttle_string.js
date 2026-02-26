export class Throttle_string {
    constructor(waitMs, callback) {
        this.waitMs = waitMs;
        this.callback = callback;
        this.buffer = '';
        this.timerId = null;
        this.lastExecTime = 0;   // 上一次真正执行的时间
    }

    update(str) {
        this.buffer += str;
        const now = Date.now();

        // 如果距离上次执行已经超过 waitMs，立刻执行
        if (now - this.lastExecTime >= this.waitMs) {
            this.fire();
        } else {
            // 否则等“剩余时间”到了再执行一次（期间重复调用只更新 buffer）
            if (!this.timerId) {
                this.timerId = setTimeout(() => this.fire(), this.waitMs - (now - this.lastExecTime));
            }
        }
    }

    fire() {
        this.callback(this.buffer);
        this.buffer = '';
        this.lastExecTime = Date.now();
        this.timerId = null;
    }

    cancel() {
        clearTimeout(this.timerId);
        this.timerId = null;
        this.buffer = '';
    }
}