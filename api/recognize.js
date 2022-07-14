const {Recognize} = require("../app");
const querystring = require("querystring");
const { v4: uuidv4 } = require("uuid");
var buffer = ''

module.exports = async (req, res) => {
    let filename = uuidv4();

    req.setEncoding("binary");
    let body = ""; // 文件数据
    let fileName = ""; // 文件名
    const boundary = req.headers["content-type"]
        .split("; ")[1]
        .replace("boundary=", "");
    req.on("data", function (chunk) {
        body += chunk;
    });
    req.on("end", function (filename) {
        const file = querystring.parse(body, "\r\n", ":");
        var fileInfo = file["Content-Disposition"].split("; ");
        const entireData = body.toString();
        contentType = file["Content-Type"].substring(1);
        const upperBoundary = entireData.indexOf(contentType) + contentType.length;
        const shorterData = entireData.substring(upperBoundary);
        const binaryDataAlmost = shorterData
            .replace(/^\s\s*/, "")
            .replace(/\s\s*$/, "");
        const binaryData = binaryDataAlmost.substring(0, binaryDataAlmost.indexOf("--" + boundary + "--"));
        buffer = new Buffer.from(binaryData, "binary");
    });

    let result = await Recognize(buffer);
    if (result?.code !== 0) {
        res.status(500)
    }
    return res.json(result);
}