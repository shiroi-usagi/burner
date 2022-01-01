# Burner

![logo](https://images.weserv.nl/?url=raw.githubusercontent.com/shiroi-usagi/burner/main/logo.png&w=64&mask=circle)

## Flags

```
      --ignore-font-error   skip font errors during encode
  -i, --input string        directory of the input files (default "./in")
  -m, --mode string         mode of the encoding
                              smp4 - Sample MP4. Encodes a sample with the subtitle burned on the video. Creates hardsub.
                              fmp4 - Fragmented MP4. Encodes a fragmented video (HLS) with the subtitle burned on the video. Creates hardsub.
                              mp4 - MP4. Encodes a video with the subtitle burned on the video. Creates hardsub.
                              transcode - Transcode. Encodes a video with the given options while keeping the original settings. Creates softsub.
  -o, --output string       directory of the output files (default "./out")
      --v-bitrate string    target video bitrate (default "1371k")
      --v-height int        target video height (default 720)
      --v-keep-bitrate      disables bitrate modification when the original file size smaller than the expected
      --v-upscaling         enable/disable upscaling
  -v, --verbose             make output verbose
```

