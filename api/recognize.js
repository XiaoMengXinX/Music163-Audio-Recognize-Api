const {Recognize} = require("../app");
const formidable = require("formidable");
const fs = require("fs");

module.exports = async (req, res) => {
    const form = new formidable.IncomingForm();
    form.parse(req, async (err, fields, files) => {
        if (err || files.audio?.filepath === undefined) {
            res.status(500).json({data: null, code: -1, message: "Parse form error"});
            return;
        }
        let songData = fs.readFileSync(files.audio.filepath);
        let result = await Recognize(songData);
        if (result?.code !== 0) {
            res.status(500)
        }
        return res.json(result);
    });
}