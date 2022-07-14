const NeteaseUtils = require('./sandbox.bundle')
const {AudioContext} = require('web-audio-api-cjs');
const axios = require('axios');
const {v4: uuidv4} = require('uuid');

const Recognize = async (songData) => {
    let result = ''
    try {
        if (songData.length < 1) {
            return {data: null, code: -1, message: 'No song data'}
        }

        const audioCtx = new AudioContext();

        try {
            let data = await audioCtx.decodeAudioData(songData)
            var encoded = await NeteaseUtils.Encode(data, 0, 6, 0)
        } catch (e) {
            console.log(e)
            return {data: null, code: -1, message: 'Decode audio data error'}
        }

        const queryData = new URLSearchParams({
            'sessionId': uuidv4(),
            'algorithmCode': 'shazam_v2',
            'duration': '6',
            'rawdata': encoded,
            'times': '2',
            'decrypt': '1'
        }).toString();

        await axios({
            method: 'post',
            url: 'https://interface.music.163.com/api/music/audio/match',
            headers: {
                "accept": "*/*",
                "accept-language": "zh-CN,zh-TW;q=0.9,zh;q=0.8",
                "cache-control": "no-cache",
                "content-type": "application/x-www-form-urlencoded; charset=UTF-8",
                "pragma": "no-cache",
                "sec-ch-ua-platform": "\"Windows\"",
                "sec-fetch-dest": "empty",
                "sec-fetch-mode": "cors",
                "sec-fetch-site": "none",
            },
            data: queryData
        }).then(function (response) {
            result = response.data
        }).catch(function (error) {
            console.log(error);
            result = {data: null, code: -1, message: 'Request error'}
        });
    } catch (e) {
        console.log(e)
        result = {data: null, code: -1, message: 'Unknown error'}
    }
    return result
}

module.exports = {
    Recognize
}