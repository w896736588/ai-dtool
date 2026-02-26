<template>
  <div ref="mainCard" style="text-align: center;">
    <div class="flex gap-4 mb-4 items-center timeTransferDiv">
      <el-input v-model="currentTime.sourceTime" class="timeInput" placeholder="当前时间戳"/>
      &nbsp;
      <el-button disabled round size="small" type="primary">反转</el-button>
      &nbsp;
      <el-input v-model="currentTime.transferResult" class="timeInput" placeholder="转换结果"/>
    </div>
    <div v-for="(value, key) in timeTransferList" :key="key" class="flex gap-4 mb-4 items-center timeTransferDiv">
      <el-input v-model="value.sourceTime" class="timeInput" placeholder="任意时间戳或字符串时间"/>
      &nbsp;
      <el-button round size="small" type="primary" @click="revert(value)">反转</el-button>
      <el-button round size="small" type="primary" @click="to235959(key ,value)">23.59.59</el-button>
      <el-button round size="small" type="primary" @click="to000000(key , value)">00.00.00</el-button>
<!--      <el-dropdown trigger="click">-->
<!--        <el-button type="primary" round size="small">-->
<!--          选项<el-icon class="el-icon&#45;&#45;right"><arrow-down /></el-icon>-->
<!--        </el-button>-->
<!--        <template #dropdown>-->
<!--          <el-dropdown-menu>-->
<!--            <el-dropdown-item @click="revert(value)">反转</el-dropdown-item>-->
<!--            <el-dropdown-item @click="to235959(value)">23.59.59</el-dropdown-item>-->
<!--            <el-dropdown-item @click="to000000(value)">00.00.00</el-dropdown-item>-->
<!--          </el-dropdown-menu>-->
<!--        </template>-->
<!--      </el-dropdown>-->
      &nbsp;
      <el-input v-model="value.transferResult" class="timeInput" placeholder="转换结果"/>
    </div>
    <el-button @click="reset" type="primary">重 置</el-button>
  </div>
</template>

<script>
import {watch} from "vue";

export default {
  props: {
    shellShowResult: {
      type: String
    },
  },
  components: {},
  data() {
    return {
      currentTime: {
        sourceTime: '',
        transferResult: '',
        color: 'black',
      },
      timeTransferList: [
        {
          sourceTime: '',
          transferResult: '',
          isError: false,
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
        {
          sourceTime: '',
          transferResult: '',
          color: 'black',
        },
      ],
    }
  },
  mounted: function () {
    this.interval()
    this.watchInit()
  },
  methods: {
    reset : function (){
      for (let i = 0; i < this.timeTransferList.length; i++) {
        this.timeTransferList[i].sourceTime = ''
        this.timeTransferList[i].transferResult = ''
      }
    },
    //定时处理当前时间戳
    interval: function () {
      let _that = this
      setInterval(function () {
        _that.currentTime.sourceTime = Math.floor(Date.now() / 1000)
      }, 1000)
    },
    watchInit: function () {
      let _that = this
      //监听当前时间
      watch(_that.currentTime, function (newValue, oldValue) {
        _that.timeChange(_that.currentTime, newValue.sourceTime)
      })
      //监听列表
      for (let i = 0; i < _that.timeTransferList.length; i++) {
        watch(_that.timeTransferList[i], function (newValue, oldValue) {
          if(_that.timeTransferList[i].sourceTime === ''){
            _that.timeTransferList[i].transferResult = '';
          }else if(_that.getTimestampByInt(_that.timeTransferList[i].sourceTime) === null &&
              _that.getTimestampByString(_that.timeTransferList[i].sourceTime) === null){
            _that.timeTransferList[i].transferResult = '';
          }else{
            _that.timeChange(_that.timeTransferList[i], newValue.sourceTime)
          }
        })
      }
    },
    timeChange: function (timeObj, newTime) {
      let _that = this
      let newTimestamp = _that.getTimestampByInt(newTime)
      if (newTimestamp !== null) {
        timeObj.transferResult = _that.formatDate(newTimestamp)
      } else {
        if(newTime.length === 8){ //处理 20220101这种格式
          const year = parseInt(newTime.substring(0, 4), 10);
          const month = parseInt(newTime.substring(4, 6), 10);
          const day = parseInt(newTime.substring(6, 8), 10);
          newTime = year + '-' +month + '-'+day;
        }
        newTimestamp = _that.getTimestampByString(newTime)
        if (newTimestamp !== null) {
          timeObj.transferResult = Math.floor(newTimestamp.getTime() / 1000)
        }
      }
    },
    revert: function (timeObj) {
      let transferResult = timeObj.transferResult
      timeObj.transferResult = timeObj.sourceTime
      timeObj.sourceTime = transferResult
    },
    to235959: function (key , timeObj) {
      let _that = this
      let newTimestamp = _that.getTimestampByString(timeObj.sourceTime)
      if (newTimestamp !== null) { //正确的字符串时间
        newTimestamp.setHours(23, 59, 59, 0)
        _that.timeTransferList[key].sourceTime = _that.formatDate(newTimestamp)
        _that.timeTransferList[key].transferResult = Math.floor(newTimestamp.getTime()/1000)
      }else{
        newTimestamp = _that.getTimestampByInt(timeObj.sourceTime)
        if (newTimestamp !== null) {
          newTimestamp.setHours(23, 59, 59, 0)
          _that.timeTransferList[key].transferResult = _that.formatDate(newTimestamp)
          _that.timeTransferList[key].sourceTime = Math.floor(newTimestamp.getTime()/1000)
        }
      }
    },
    to000000 : function (key , timeObj){
      let _that = this
      let newTimestamp = _that.getTimestampByString(timeObj.sourceTime)
      if (newTimestamp !== null) { //正确的时间
        newTimestamp.setHours(0, 0, 0, 0)
        _that.timeTransferList[key].sourceTime = _that.formatDate(newTimestamp)
        _that.timeTransferList[key].transferResult = Math.floor(newTimestamp.getTime()/1000)
      }else{
        newTimestamp = _that.getTimestampByInt(timeObj.sourceTime)
        if (newTimestamp !== null) {
          newTimestamp.setHours(0, 0, 0, 0)
          _that.timeTransferList[key].transferResult = _that.formatDate(newTimestamp)
          _that.timeTransferList[key].sourceTime = Math.floor(newTimestamp.getTime()/1000)
        }
      }
    },
    //通过时间戳拿到时间
    getTimestampByInt: function (stringTimestamp) {
      stringTimestamp = stringTimestamp + '' //转为字符串
      if (stringTimestamp.length === 13) { //毫秒级时间戳
        stringTimestamp = stringTimestamp.substring(0, 10)
      } else if (stringTimestamp.length !== 10) {
        return null
      }
      let timestamp = parseInt(stringTimestamp, 10);
      if (timestamp.toString() !== stringTimestamp) {
        //非时间戳或毫秒时间戳
        return null
      }
      return new Date(timestamp * 1000);
    },
    //通过字符串获取时间
    getTimestampByString: function (timeString) {
      let regex = /^[0-9]{4}[-\/ ]?[0-9]{1,2}[-\/ ]?[0-9]{1,2}\s*([0-9]{1,2}[-\/: ]{1}[0-9]{1,2}[-\/: ]{1}[0-9]{1,2})*$/;
      if(!regex.test(timeString)){
        return null
      }
      // 提取日期部分
      const dateParts = timeString.match(/^([0-9]{4})[-\/ ]?([0-9]{1,2})[-\/ ]?([0-9]{1,2})/);
      const [ , year, month, day ] = dateParts || [];

      //移除年月日部分
      const dateEndIndex = (dateParts && dateParts.index + dateParts[0].length) || 0;
      const timeStringRemainder = timeString.slice(dateEndIndex).trim();

      // 提取时间部分（如果有）
      const timeParts = timeStringRemainder.match(/([0-9]{1,2})[-\/: ]([0-9]{1,2})[-\/: ]([0-9]{1,2})/g);
      if (timeParts) {
        let rHour = ''
        let rMinute = ''
        let rSecond = ''
        timeParts.forEach(timePart => {
          const [ , hour, minute, second ] = timePart.match(/([0-9]{1,2})[-\/: ]([0-9]{1,2})[-\/: ]([0-9]{1,2})/);
          rHour = hour
          rMinute = minute
          rSecond = second
        });
        return new Date(year,month-1,day,rHour,rMinute,rSecond,0);
      }else{
        return new Date(year,month-1,day,0,0,0,0);
      }
    },
    //格式化输出
    formatDate: function (date) {
      const year = date.getFullYear();
      const month = ('0' + (date.getMonth() + 1)).slice(-2);
      const day = ('0' + date.getDate()).slice(-2);
      const hours = ('0' + date.getHours()).slice(-2);
      const minutes = ('0' + date.getMinutes()).slice(-2);
      const seconds = ('0' + date.getSeconds()).slice(-2);
      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    }
  }
}
</script>

<style>
.timeTransferDiv {
  margin: 15px;
}

.timeInput {
  width: 340px
}

.timeInputErr {
  width: 340px;
  color: red;
}

.example-showcase .el-dropdown-link {
  cursor: pointer;
  color: var(--el-color-primary);
  display: flex;
  align-items: center;
}
</style>
