const {Recognize} = require("../app");
const querystring = require("querystring");

module.exports = async (req, res) => {
    req.setEncoding("binary");
    let body = "";
    const boundary = req.headers["content-type"]
        .split("; ")[1]
        .replace("boundary=", "");
    req.on("data", function (chunk) {
        body += chunk;
    });
    req.on("end", async function () {
        let bufferData = ''
        try {
            const file = querystring.parse(body, "\r\n", ":");
            const entireData = body.toString();
            let contentType = file["Content-Type"].substring(1);
            const upperBoundary = entireData.indexOf(contentType) + contentType.length;
            const shorterData = entireData.substring(upperBoundary);
            const binaryDataAlmost = shorterData
                .replace(/^\s\s*/, "")
                .replace(/\s\s*$/, "");
            const binaryData = binaryDataAlmost.substring(0, binaryDataAlmost.indexOf("--" + boundary + "--"));
            bufferData = new Buffer.from(binaryData, "binary");
        } catch(e) {
            console.log(e)
        }

        if (bufferData.length < 1) {
            res.status(500)
            res.json({data: null, code: -1, message: 'Failed to parse song data'});
            return
        }

        let result = await Recognize(bufferData);
        if (result?.code !== 200) {
            res.status(500)
        }
        res.json(result);
    });
}