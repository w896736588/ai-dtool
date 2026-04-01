const assert = require('assert')

const {
  normalizeSupervisorProgramNameFromHeader,
  parseSupervisorStatusLine,
} = require('../src/utils/supervisor_status.cjs')

const run = () => {
  assert.strictEqual(
    normalizeSupervisorProgramNameFromHeader('[program:aiServer]'),
    'aiServer',
    'program 头应解析成 supervisor 进程名'
  )

  assert.strictEqual(
    normalizeSupervisorProgramNameFromHeader('[group:clockInConsumer]'),
    'clockInConsumer',
    'group 头也应解析成 supervisor 进程名前缀'
  )

  assert.strictEqual(
    normalizeSupervisorProgramNameFromHeader(' [fcgi-program:message_service] '),
    'message_service',
    '其他带冒号的 supervisor 头也应提取冒号后的名字'
  )

  assert.strictEqual(
    parseSupervisorStatusLine('ec xkf_common5 supervisorctl status'),
    null,
    '命令回显行不应被识别为进程状态'
  )

  assert.deepStrictEqual(
    parseSupervisorStatusLine('aiServer                                           RUNNING   pid 89, uptime 195 days, 2:36:10\r'),
    {
      processName: 'aiServer',
      groupName: 'aiServer',
      statusText: 'RUNNING pid 89, uptime 195 days, 2:36:10',
    },
    '普通 RUNNING 行应被正确解析'
  )

  assert.deepStrictEqual(
    parseSupervisorStatusLine('clockInConsumer:clockInConsumer_00                 RUNNING   pid 83, uptime 195 days, 2:36:10'),
    {
      processName: 'clockInConsumer:clockInConsumer_00',
      groupName: 'clockInConsumer',
      statusText: 'RUNNING pid 83, uptime 195 days, 2:36:10',
    },
    '组进程状态行应保留原始进程名并提取组名前缀'
  )

  assert.deepStrictEqual(
    parseSupervisorStatusLine('message_service                                    FATAL     can\'t find command \'/var/www/apps/message_service/message_service\''),
    {
      processName: 'message_service',
      groupName: 'message_service',
      statusText: 'FATAL can\'t find command \'/var/www/apps/message_service/message_service\'',
    },
    'FATAL 行应被正确解析'
  )

  console.log('supervisor_status_parser tests passed')
}

run()
