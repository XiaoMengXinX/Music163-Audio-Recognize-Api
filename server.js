const express = require('express')
const api = require('./api/recognize');

var expressApp = express();
expressApp.use('/api/recognize', api);
expressApp.listen(8000, () => {
    console.log('Server started on port 8000');
})

