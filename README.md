# Burner

![logo](https://images.weserv.nl/?url=raw.githubusercontent.com/shiroi-usagi/burner/main/logo.png&w=64&mask=circle)

## Flags

| flag                 | values                            | default value  | description                   |
|----------------------|-----------------------------------|----------------|-------------------------------|
| `-ignore-font-error` | boolean                           | `false`        | Ignore font errors.           |
| `-input`             | string                            | `./in`         | Folder of the input files.    |
| `-mode`              | `smp4`, `fmp4`, `mp4`, `transcode`| none*          | Name of the transcode preset. |
| `-output`            | string                            | `./out`        | Folder of the output files.   |
| `-v`                 | boolean                           | `false`        | Make output verbose.          |
| `-v-bitrate`         | formatted bitrate                 | `1371k`        | Target video bitrate.         |
| `-v-height`          | integer                           | `720`          | Target video height.          |
| `-v-keep-bitrate`    | boolean                           | `false`        | Disable bitrate modification. |
| `-v-upscaling`       | boolean                           | `false`        | Enable/disable upscaling.     |

> `*` If no value is set then the app will request it. _(Interactive CLI)_

## Binaries

On Windows if `ffmpeg` or `ffprobe` are not available on `PATH` the application will download the "essential" build from https://github.com/GyanD/codexffmpeg/releases/tag/4.4.
