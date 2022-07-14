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
        const file = querystring.parse(body, "\r\n", ":");
        const entireData = body.toString();
        let contentType = file["Content-Type"].substring(1);
        const upperBoundary = entireData.indexOf(contentType) + contentType.length;
        const shorterData = entireData.substring(upperBoundary);
        const binaryDataAlmost = shorterData
            .replace(/^\s\s*/, "")
            .replace(/\s\s*$/, "");
        const binaryData = binaryDataAlmost.substring(0, binaryDataAlmost.indexOf("--" + boundary + "--"));
        const bufferData = new Buffer.from(binaryData, "binary");

        let result = await Recognize(bufferData);
        if (result?.code !== 0) {
            res.status(500)
        }
        res.json(result);
    });
}