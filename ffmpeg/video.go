package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func filename(path string) string {
	return filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
}

type ffmpegOption struct {
	firstPass  bool
	secondPass bool

	flag  string
	value string
}

type Transcoder struct {
	executable string
	input      string
	outFile    string
	outDir     string

	options []ffmpegOption
}

// NewTranscoder builds a Transcoder for fragmented mp4 with preset data
func NewTranscoder(executable, input, outDir, bitrate string, f Filter) *Transcoder {
	t := Transcoder{
		executable: executable,

		input:   input,
		outFile: filepath.Base(input),
		outDir:  outDir,
	}
	t.VideoBitrate(bitrate)
	t.Tune("animation")
	t.Preset("medium")
	t.PixelFormat("yuv420p")
	// Disable subtitle burning in this preset
	f.Subtitle = ""
	t.Filter(f)
	t.AudioCodec("aac")
	t.AudioBitrate("128k")
	t.SubtitleCodec("copy")
	// Keep audio
	t.Map("0:a")
	// Keep video
	t.Map("0:s")
	// https://superuser.com/questions/1320275/map-t-vs-map-0t-vs-tcodec-copy
	t.Map("0:t")
	t.AttachmentCopy()
	return &t
}

// NewFragmentedMp4Transcoder builds a Transcoder for fragmented mp4 with preset data
func NewFragmentedMp4Transcoder(executable, input, outDir, bitrate string, f Filter) *Transcoder {
	t := Transcoder{
		executable: executable,

		input:   input,
		outFile: "out.m3u8",
		outDir:  filepath.Join(outDir, filename(input)),
	}
	t.VideoCodec("libx264")
	t.VideoBitrate(bitrate)
	t.Tune("animation")
	t.Preset("medium")
	t.PixelFormat("yuv420p")
	t.Filter(f)
	t.AudioCodec("aac")
	t.AudioBitrate("128k")
	t.AudioChannels("2")
	t.HlsFlags("append_list")
	t.HlsTime(10 * time.Second)
	t.HlsListSize(0)
	t.HlsSegmentType("fmp4")
	t.SkipSubtitleStream()
	return &t
}

// NewMp4Transcoder builds a Transcoder for mp4 with preset data
func NewMp4Transcoder(executable, input, outDir, bitrate string, f Filter) *Transcoder {
	t := Transcoder{
		executable: executable,

		input:   input,
		outFile: fmt.Sprintf("%s.mp4", filename(input)),
		outDir:  outDir,
	}
	t.VideoCodec("libx264")
	t.VideoBitrate(bitrate)
	t.Tune("animation")
	t.Preset("medium")
	t.PixelFormat("yuv420p")
	t.Filter(f)
	t.AudioCodec("aac")
	t.AudioBitrate("128k")
	t.AudioChannels("2")
	t.SkipSubtitleStream()
	return &t
}

// NewSampleMp4Transcoder builds a Transcoder for fragmented mp4 with preset data
func NewSampleMp4Transcoder(executable, input, outDir, bitrate string, f Filter) *Transcoder {
	t := Transcoder{
		executable: executable,

		input:   input,
		outFile: fmt.Sprintf("%s_sample.mp4", filename(input)),
		outDir:  outDir,
	}
	t.Seek(time.Minute)
	t.Duration(time.Minute)
	t.VideoCodec("libx264")
	t.VideoBitrate(bitrate)
	t.Tune("animation")
	t.Preset("medium")
	t.PixelFormat("yuv420p")
	t.Filter(f)
	t.AudioCodec("aac")
	t.AudioBitrate("128k")
	t.AudioChannels("2")
	t.SkipSubtitleStream()
	return &t
}

// AudioChannels downmux the output channels to the specified value.
func (t *Transcoder) AudioChannels(c string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-ac", value: c,
	})
}

// OutDir is the target directory of the output
func (t *Transcoder) OutDir() string {
	return t.outDir
}

// VideoCodec sets the codec for all video streams
func (t *Transcoder) VideoCodec(c string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-c:v", value: c,
	})
}

// VideoBitrate sets the bitrate for all video streams
func (t *Transcoder) VideoBitrate(b string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-b:v", value: b,
	})
}

// Tune sets the tune settings for the encoding
func (t *Transcoder) Tune(tune string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-tune", value: tune,
	})
}

// Preset sets the preset value for the encoding
func (t *Transcoder) Preset(p string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-preset", value: p,
	})
}

// PixelFormat sets the pixel format for the encoding
func (t *Transcoder) PixelFormat(pf string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-pix_fmt", value: pf,
	})
}

func (t *Transcoder) Filter(f Filter) {
	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-filter_complex", value: f.String(),
	})
}

// AudioCodec sets the codec for all audio streams
func (t *Transcoder) AudioCodec(c string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-c:a", value: c,
	})
}

// AudioBitrate sets the bitrate for all audio streams
func (t *Transcoder) AudioBitrate(b string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-b:a", value: b,
	})
}

// SubtitleCodec sets the codec for all subtitle streams
func (t *Transcoder) SubtitleCodec(c string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-c:s", value: c,
	})
}

// SkipSubtitleStream sets a flag to skip inclusion of subtitle streams
func (t *Transcoder) SkipSubtitleStream() {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-sn",
	})
}

// Map maps specific streams to the target file
func (t *Transcoder) Map(v string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-map", value: v,
	})
}

// AttachmentCopy copies all attachment streams to the output
func (t *Transcoder) AttachmentCopy() {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-c:t", value: "copy",
	})
}

// HlsFlags sets the `-hls_flags` option for the encoding
//
// Possible values:
//
// `single_file`
// If this flag is set, the muxer will store all segments in a single MPEG-TS file, and will use byte ranges in
// the playlist. HLS playlists generated with this way will have the version number 4. For example:
//
// ```
// ffmpeg -i in.nut -hls_flags single_file out.m3u8
// ```
// Will produce the playlist, out.m3u8, and a single segment file, out.ts.
//
// `delete_segments`
// Segment files removed from the playlist are deleted after a period of time equal to the duration of the segment
// plus the duration of the playlist.
//
// `append_list`
// Append new segments into the end of old segment list, and remove the `#EXT-X-ENDLIST` from the old segment list.
//
// `round_durations`
// Round the duration info in the playlist file segment info to integer values, instead of using floating point.
//
// `discont_start`
// Add the `#EXT-X-DISCONTINUITY` tag to the playlist, before the first segment`s information.
//
// `omit_endlist`
// Do not append the EXT-X-ENDLIST tag at the end of the playlist.
//
// `periodic_rekey`
// The file specified by hls_key_info_file will be checked periodically and detect updates to the encryption info.
// Be sure to replace this file atomically, including the file containing the AES encryption key.
//
// `independent_segments`
// Add the `#EXT-X-INDEPENDENT-SEGMENTS` to playlists that has video segments and when all the segments of that playlist
// are guaranteed to start with a Key frame.
//
// `iframes_only`
// Add the `#EXT-X-I-FRAMES-ONLY` to playlists that has video segments and can play only I-frames in the
// `#EXT-X-BYTERANGE` mode.
//
// `split_by_time`
// Allow segments to start on frames other than keyframes. This improves behavior on some players when the time
// between keyframes is inconsistent, but may make things worse on others, and can cause some oddities during seeking. This flag should be used with the hls_time option.
//
// `program_date_time`
// Generate EXT-X-PROGRAM-DATE-TIME tags.
//
// `second_level_segment_index`
// Makes it possible to use segment indexes as %%d in hls_segment_filename expression besides date/time values when
// strftime is on. To get fixed width numbers with trailing zeroes, %%0xd format is available where x is the required
// width.
//
// `second_level_segment_size`
// Makes it possible to use segment sizes (counted in bytes) as %%s in hls_segment_filename expression besides
// date/time values when strftime is on. To get fixed width numbers with trailing zeroes, %%0xs format is available
// where x is the required width.
//
// `second_level_segment_duration`
// Makes it possible to use segment duration (calculated in microseconds) as %%t in hls_segment_filename expression
// besides date/time values when strftime is on. To get fixed width numbers with trailing zeroes, %%0xt format is
// available where x is the required width.
//
// ```
// ffmpeg -i sample.mpeg \
//   -f hls -hls_time 3 -hls_list_size 5 \
//   -hls_flags second_level_segment_index+second_level_segment_size+second_level_segment_duration \
//   -strftime 1 -strftime_mkdir 1 -hls_segment_filename "segment_%Y%m%d%H%M%S_%%04d_%%08s_%%013t.ts" stream.m3u8
// ```
// This will produce segments like this: segment_20170102194334_0003_00122200_0000003000000.ts,
// segment_20170102194334_0004_00120072_0000003000000.ts etc.
//
// `temp_file`
// Write segment data to filename.tmp and rename to filename only once the segment is complete. A webserver serving
// up segments can be configured to reject requests to *.tmp to prevent access to in-progress segments before they
// have been added to the m3u8 playlist. This flag also affects how m3u8 playlist files are created. If this flag is
// set, all playlist files will written into temporary file and renamed after they are complete, similarly as segments
// are handled. But playlists with file protocol and with type (hls_playlist_type) other than vod are always written
// into temporary file regardless of this flag. Master playlist files (master_pl_name), if any, with file protocol,
// are always written into temporary file regardless of this flag if master_pl_publish_rate value is other than zero.
func (t *Transcoder) HlsFlags(f string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-hls_flags", value: f,
	})
}

// HlsTime sets the `-hls_time` option for the encoding
//
// Set the target segment length. Default value is 2. Segment will be cut on the next key frame after
// this time has passed.
func (t *Transcoder) HlsTime(d time.Duration) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-hls_time", value: fmt.Sprintf("%.0f", d.Seconds()),
	})
}

// HlsListSize sets the `-hls_list_size` option for the encoding
//
// Set the maximum number of playlist entries. If set to 0 the list file will contain all the segments.
// Default value is 5.
func (t *Transcoder) HlsListSize(ls uint) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-hls_list_size", value: fmt.Sprintf("%d", ls),
	})
}

// HlsSegmentType sets the `-hls_segment` option for the encoding
//
// Possible values:
//
// `mpegts`
// Output segment files in MPEG-2 Transport Stream format. This is compatible with all HLS versions.
//
// `fmp4`
// Output segment files in fragmented MP4 format, similar to MPEG-DASH. fmp4 files may be used in HLS version 7 and above.
func (t *Transcoder) HlsSegmentType(st string) {
	t.options = append(t.options, ffmpegOption{
		firstPass: false, secondPass: true, flag: "-hls_segment_type", value: st,
	})
}

// Seek sets the `-ss` option for the encoding
//
// When used as an input option (before `-i`), seeks in this input file to position. Note that in most formats
// it is not possible to seek exactly, so `ffmpeg` will seek to the closest seek point before position. When
// transcoding and -accurate_seek is enabled (the default), this extra segment between the seek point and
// position will be decoded and discarded. When doing stream copy or when -noaccurate_seek is used, it will
// be preserved.
//
// When used as an output option (before an output url), decodes but discards input until the timestamps reach position.
func (t *Transcoder) Seek(p time.Duration) {
	unix := time.Unix(0, 0).Add(p).UTC()

	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-ss", value: unix.Format("15:04:05"),
	})
}

// Duration sets the `-t` option for the encoding
//
// When used as an input option (before `-i`), limit the duration of data read from the input file.
//
// When used as an output option (before an output url), stop writing the output after its duration reaches duration.
func (t *Transcoder) Duration(d time.Duration) {
	unix := time.Unix(0, 0).Add(d).UTC()

	t.options = append(t.options, ffmpegOption{
		firstPass: true, secondPass: true, flag: "-t", value: unix.Format("15:04:05"),
	})
}

func (t Transcoder) FirstPass() *exec.Cmd {
	var args []string
	args = append(args, "-y")          // Overwrite output files without asking.
	args = append(args, "-i", t.input) // Input file url
	args = append(args, "-pass", "1")  // Select the pass number 1
	for _, option := range t.options {
		if !option.firstPass {
			continue
		}
		if option.value == "" {
			args = append(args, option.flag)
		} else {
			args = append(args, option.flag, option.value)
		}
	}
	args = append(args, "-an")       // Skip inclusion of audio
	args = append(args, "-f", "mp4") // Force output file format
	args = append(args, os.DevNull)  // Set output to null
	cmd := exec.Command(t.executable, args...)
	cmd.Dir = t.outDir
	return cmd
}

func (t Transcoder) SecondPass() *exec.Cmd {
	var args []string
	args = append(args, "-i", t.input) // Input file url
	args = append(args, "-pass", "2")  // Select the pass number 1
	for _, option := range t.options {
		if !option.secondPass {
			continue
		}
		if option.value == "" {
			args = append(args, option.flag)
		} else {
			args = append(args, option.flag, option.value)
		}
	}
	args = append(args, t.outFile) // Set output to null
	cmd := exec.Command(t.executable, args...)
	cmd.Dir = t.outDir
	return cmd
}

type Filter struct {
	// Source file for subtitle
	Subtitle string
	// Width value of the scale filter
	Width int
	// Height value of the scale filter
	Height int
	// Enable/disable upscaling in scale filter with use of min
	Upscaling bool
}

func (f Filter) String() string {
	var filters []string
	if f.Subtitle != "" {
		// path must be escaped for -vf and -filter_complex
		p := f.Subtitle
		p = strings.ReplaceAll(p, `\`, `\\`)
		p = strings.ReplaceAll(p, `:`, `\:`)
		filters = append(filters, fmt.Sprintf(`subtitles='%s'`, p))
	}
	if f.Width != 0 || f.Height != 0 {
		if f.Upscaling {
			filters = append(filters, fmt.Sprintf("scale=%d:%d", f.Width, f.Height))
		} else {
			filters = append(filters, fmt.Sprintf("scale='min(%d,iw)':'min(%d,ih)'", f.Width, f.Height))
		}
	}
	return strings.Join(filters, ", ")
}

func BitrateToKilobit(bitrate string) int64 {
	switch {
	case strings.HasSuffix(bitrate, "k"):
		i, err := strconv.ParseInt(bitrate[:len(bitrate)-1], 10, 64)
		if err != nil {
			panic(err)
		}
		return i
	case strings.HasSuffix(bitrate, "M"):
		i, err := strconv.ParseInt(bitrate[:len(bitrate)-1], 10, 64)
		if err != nil {
			panic(err)
		}
		return i * 1024
	}
	return 0
}

func KilobitToBitrate(kilobit int64) string {
	return fmt.Sprintf("%dk", kilobit)
}
