module.exports = {
    devServer: {
        proxy: {
            '/buyer': {
                target: 'http://localhost:8081',
                changeOrigin: true,
            },
        },
    },
};
