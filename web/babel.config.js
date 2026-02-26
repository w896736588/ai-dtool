module.exports = {
    presets: [
        '@vue/cli-plugin-babel/preset'
    ],
    plugins: [
        '@babel/plugin-transform-private-methods',
        // 如果还报私有字段错误，再加：
        '@babel/plugin-transform-class-properties'
    ]
}
