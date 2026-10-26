# NeteaseCloudMusic-Audio-Recognize

## Usage

```
$ npm i
$ npm run test-ts samples/Komplexe.mp3

> ncm-audio-recognize@1.0.0 test
> tsx test.ts "samples/Komplexe.mp3"

Module loading: /Users/natsuki/workspace/ncm-audio-recognize/afp.wasm
Module loaded: [object WebAssembly.Instance]
Encoded Data:  W5/Hd3yYCY2E....
[1] Komplexe - 雄之助 Album: 「maimai でらっくす ベストアルバムちほー」 CID: [1914659931]
```

## Usage (with CommonJS)

> We refactor the `web-audio-api` project to use CommonJS
> The CommonJS module name is `web-audio-api-cjs`

Try it with command:

```
npm run test-js samples/Komplexe.mp3
```

## Documentation

```NeteaseUtils.Encode(audiodata, from, len, channel)```

| Params          | Type          | Description                          |
|-----------------|---------------|--------------------------------------|
| ```audiodata``` | `AudioBuffer` | The audio data.                      |
| ```from```      | `number`      | The time when the audio data started |
| ```len```       | `number`      | The secs of the audio data.          |
| ```channel```   | `number`      | The channel of the audio data.       |


## Go package

`github.com/XiaoMengXinX/Music163-Audio-Recognize-Api` (package `ncmrecognize`) is a pure-Go port: the original `afp.wasm`
fingerprint algorithm runs in-process via [wazero](https://github.com/tetratelabs/wazero) (no cgo, no Node.js).
Its output is byte-identical to the Node implementation.

```go
import ncm "github.com/XiaoMengXinX/Music163-Audio-Recognize-Api"

resp, err := new(ncm.Client).RecognizeFile(ctx, "song.ogg") // wav / mp3 / flac / ogg (vorbis, opus)
fmt.Println(string(resp.Raw))                                // raw API JSON; resp.Code, resp.Data too

// Already have PCM? 
fp, _ := ncm.Fingerprint(ctx, pcm, 44100, 0) // mono float32 samples, offset in seconds
resp, _ = new(ncm.Client).Match(ctx, fp)
```

CLI: `go run ./cmd/ncmrecognize song.ogg`

Supported input (pure Go decoders, detected by content): WAV (PCM), MP3, FLAC, Ogg Vorbis, Ogg Opus.
AAC/M4A/WMA are not supported; decode them yourself and pass PCM to `Fingerprint` / `Client.RecognizePCM`.

How it was reverse engineered: `afp.wasm` is an Emscripten/embind build exposing `ExtractQueryFP(std::string) -> vector<char>`.
The package patches extra exports onto the wasm (embind raw invokers + destructor), captures the member-function
pointers from the embind registration calls, and calls the invokers directly, so no embind runtime is needed.
The input is 6 s of channel 0, point-sampled to 8 kHz float32.

## Credits

[MicroCBer/BeatSaverSongData](https://github.com/MicroCBer/BeatSaverSongData)