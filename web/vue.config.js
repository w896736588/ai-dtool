const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: true,
  lintOnSave: false,
  pages: {
    index: {
      entry: 'src/main.js',
      template: 'public/index.html',
      filename: 'index.html',
      title: 'dtool',
    },
    recorder: {
      entry: 'src/components/e2e/recorder-runtime/index.js',
      template: 'src/components/e2e/recorder-runtime/proxy.html',
      filename: 'e2e-recorder.html',
      chunks: ['chunk-vendors', 'recorder'],
      title: 'recorder',
    },
  },
})
