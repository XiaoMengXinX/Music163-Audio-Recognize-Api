const fs = require("fs");
const {formidable} = require("formidable");
const {Recognize} = require("../app");

module.exports = async (req, res) => {
    let bufferData = Buffer.alloc(0);
    try {
        const [, files] = await formidable({}).parse(req);
        const file = Object.values(files).flat()[0];
        if (file) {
            bufferData = fs.readFileSync(file.filepath);
            fs.unlink(file.filepath, () => {});
        }
    } catch (e) {
        console.log(e);
    }

    if (bufferData.length < 1) {
        res.status(500);
        res.json({data: null, code: -1, message: 'Failed to parse song data'});
        return;
    }

    let result = await Recognize(bufferData);
    if (result?.code !== 200) {
        res.status(500);
    }
    res.json(result);
};
