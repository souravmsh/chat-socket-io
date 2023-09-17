module.exports = {
    entry: './app.js',
    mode: 'production',
    output: {
      path: `${__dirname}/dist`,
      filename: 'bundle.min.js',
    },
    resolve: {
        fallback: {
            "fs": false,
            "tls": false,
            "net": false,
            "path": false,
            "zlib": false,
            "http": false,
            "https": false,
            "stream": false,
            "url": false,
            "querystring": false,
            "crypto": false,
            "timers": false,
            "util": false,
            "buffer": false,
            // "crypto-browserify": require.resolve('crypto-browserify')
        },
    },
  };