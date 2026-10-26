// Command ncmrecognize identifies a song file (wav, mp3, flac, ogg) via NetEase Cloud Music.
package main

import (
	"context"
	"fmt"
	"os"

	ncm "github.com/XiaoMengXinX/Music163-Audio-Recognize-Api"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: ncmrecognize <audio.mp3|audio.wav>")
		os.Exit(2)
	}
	resp, err := new(ncm.Client).RecognizeFile(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(string(resp.Raw))
}
