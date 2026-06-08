package local

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
	"github.com/mewkiz/flac"
	"github.com/tcolgate/mp3"
)

func formatTrackDuration(seconds float64) string {
	if seconds <= 0 || !isFinite(seconds) {
		return ""
	}
	total := int(seconds + 0.5)
	if total <= 0 {
		return ""
	}
	minutes := total / 60
	secs := total % 60
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func isFinite(v float64) bool {
	return v == v && v < 1e9 && v > -1e9
}

// DurationSecondsFromMetadata reads common duration fields from tag metadata (e.g. ID3 TLEN).
func DurationSecondsFromMetadata(metadata tag.Metadata) float64 {
	if metadata == nil {
		return 0
	}
	raw := metadata.Raw()
	for _, key := range []string{"Length", "TLEN", "TLE", "duration", "©dur"} {
		if v, ok := raw[key]; ok {
			if sec := parseTagDurationSeconds(v); sec > 0 {
				return sec
			}
		}
	}
	return 0
}

func parseTagDurationSeconds(v interface{}) float64 {
	switch value := v.(type) {
	case string:
		text := strings.TrimSpace(value)
		if text == "" {
			return 0
		}
		if ms, err := strconv.ParseFloat(text, 64); err == nil && ms > 0 {
			// ID3 TLEN / Length is milliseconds.
			return ms / 1000
		}
	case int:
		if value > 0 {
			return float64(value) / 1000
		}
	case int64:
		if value > 0 {
			return float64(value) / 1000
		}
	case float64:
		if value > 0 {
			return value / 1000
		}
	}
	return 0
}

// ProbeAudioDurationFromFile estimates track length using an already-open file.
// The reader is rewound to the start when needed.
func ProbeAudioDurationFromFile(r io.ReadSeeker, ext string, fileSize int64) float64 {
	switch strings.ToLower(ext) {
	case ".mp3":
		if seconds, err := mp3DurationFast(r, fileSize); err == nil {
			return seconds
		}
		if seconds, err := mp3DurationByFullDecode(r); err == nil {
			return seconds
		}
	case ".flac":
		if seconds, err := flacDurationFromSeeker(r); err == nil {
			return seconds
		}
	case ".wav":
		if seconds, err := wavDurationFromSeeker(r); err == nil {
			return seconds
		}
	case ".ogg":
		if seconds, err := oggDurationFromSeeker(r, fileSize); err == nil {
			return seconds
		}
	}
	return 0
}

func mp3DurationFast(r io.ReadSeeker, fileSize int64) (float64, error) {
	if fileSize <= 0 {
		return 0, fmt.Errorf("invalid file size")
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	leading, err := id3v2LeadingBytes(r)
	if err != nil {
		leading = 0
	}
	if _, err := r.Seek(leading, io.SeekStart); err != nil {
		return 0, err
	}

	decoder := mp3.NewDecoder(r)
	var frame mp3.Frame
	skipped := 0
	if err := decoder.Decode(&frame, &skipped); err != nil {
		return 0, err
	}

	audioStart := leading + int64(skipped)
	frameSize := frame.Size()
	if frameSize <= 0 {
		return 0, fmt.Errorf("invalid mp3 frame size")
	}

	if _, err := r.Seek(audioStart, io.SeekStart); err != nil {
		return 0, err
	}
	frameBuf := make([]byte, frameSize)
	if _, err := io.ReadFull(r, frameBuf); err != nil {
		return 0, err
	}

	sampleRate := int(frame.Header().SampleRate())
	samplesPerFrame := frame.Samples()
	if sampleRate > 0 && samplesPerFrame > 0 {
		if seconds, ok := xingDurationSeconds(frameBuf, sampleRate, samplesPerFrame); ok {
			return seconds, nil
		}
	}

	bitrate := int(frame.Header().BitRate())
	if bitrate <= 0 {
		return 0, fmt.Errorf("invalid mp3 bitrate")
	}

	audioBytes := fileSize - audioStart
	if id3v1TrailerSize(r, fileSize) {
		audioBytes -= 128
	}
	if audioBytes <= 0 {
		return 0, fmt.Errorf("no mp3 audio payload")
	}
	return float64(audioBytes) * 8 / float64(bitrate), nil
}

func mp3DurationByFullDecode(r io.ReadSeeker) (float64, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	decoder := mp3.NewDecoder(r)
	var frame mp3.Frame
	skipped := 0
	var totalSeconds float64
	for {
		if err := decoder.Decode(&frame, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		totalSeconds += frame.Duration().Seconds()
	}
	if totalSeconds <= 0 {
		return 0, fmt.Errorf("mp3 duration is zero")
	}
	return totalSeconds, nil
}

func xingDurationSeconds(frame []byte, sampleRate, samplesPerFrame int) (float64, bool) {
	idx := bytes.Index(frame, []byte("Xing"))
	if idx < 0 {
		idx = bytes.Index(frame, []byte("Info"))
	}
	if idx < 0 || idx+8 > len(frame) {
		return 0, false
	}

	flags := binary.BigEndian.Uint32(frame[idx+4 : idx+8])
	if flags&0x01 == 0 {
		return 0, false
	}
	pos := idx + 8
	if pos+4 > len(frame) {
		return 0, false
	}
	frames := binary.BigEndian.Uint32(frame[pos : pos+4])
	if frames == 0 || sampleRate <= 0 || samplesPerFrame <= 0 {
		return 0, false
	}
	return float64(frames) * float64(samplesPerFrame) / float64(sampleRate), true
}

func id3v2LeadingBytes(r io.ReadSeeker) (int64, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	var header [10]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return 0, err
	}
	if string(header[0:3]) != "ID3" {
		return 0, nil
	}
	return 10 + int64(id3v2SyncSafeSize(header[6:10])), nil
}

func id3v2SyncSafeSize(b []byte) int {
	return int(b[0]&0x7f)<<21 | int(b[1]&0x7f)<<14 | int(b[2]&0x7f)<<7 | int(b[3]&0x7f)
}

func id3v1TrailerSize(r io.ReadSeeker, fileSize int64) bool {
	if fileSize < 128 {
		return false
	}
	if _, err := r.Seek(fileSize-128, io.SeekStart); err != nil {
		return false
	}
	var tagHeader [3]byte
	if _, err := io.ReadFull(r, tagHeader[:]); err != nil {
		return false
	}
	return string(tagHeader[:]) == "TAG"
}

func flacDurationFromSeeker(r io.ReadSeeker) (float64, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	stream, err := flac.Parse(r)
	if err != nil {
		return 0, err
	}
	if stream.Info.SampleRate == 0 || stream.Info.NSamples == 0 {
		return 0, fmt.Errorf("invalid flac stream info")
	}
	return float64(stream.Info.NSamples) / float64(stream.Info.SampleRate), nil
}

func wavDurationFromSeeker(r io.ReadSeeker) (float64, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	var riff [12]byte
	if _, err := io.ReadFull(r, riff[:]); err != nil {
		return 0, err
	}
	if string(riff[0:4]) != "RIFF" || string(riff[8:12]) != "WAVE" {
		return 0, fmt.Errorf("not a wav file")
	}

	var byteRate uint32
	var dataSize uint32
	foundFmt := false
	foundData := false

	for {
		var chunkHeader [8]byte
		if _, err := io.ReadFull(r, chunkHeader[:]); err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		switch chunkID {
		case "fmt ":
			if chunkSize < 16 {
				return 0, fmt.Errorf("invalid wav fmt chunk")
			}
			var fmtChunk [16]byte
			if _, err := io.ReadFull(r, fmtChunk[:]); err != nil {
				return 0, err
			}
			byteRate = binary.LittleEndian.Uint32(fmtChunk[8:12])
			foundFmt = true
			if chunkSize > 16 {
				if _, err := r.Seek(int64(chunkSize-16), io.SeekCurrent); err != nil {
					return 0, err
				}
			}
		case "data":
			dataSize = chunkSize
			foundData = true
			if _, err := r.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return 0, err
			}
		default:
			if _, err := r.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return 0, err
			}
		}

		if chunkSize%2 == 1 {
			if _, err := r.Seek(1, io.SeekCurrent); err != nil {
				return 0, err
			}
		}

		if foundFmt && foundData {
			break
		}
	}

	if !foundFmt || !foundData || byteRate == 0 {
		return 0, fmt.Errorf("wav missing fmt/data")
	}
	return float64(dataSize) / float64(byteRate), nil
}

func oggDurationFromSeeker(r io.ReadSeeker, fileSize int64) (float64, error) {
	if fileSize <= 0 {
		return 0, fmt.Errorf("invalid file size")
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	headSize := int64(96 * 1024)
	if fileSize < headSize {
		headSize = fileSize
	}
	head := make([]byte, headSize)
	if _, err := io.ReadFull(r, head); err != nil {
		return 0, err
	}
	sampleRate := parseOggSampleRateFromStart(head)
	if sampleRate <= 0 {
		return 0, fmt.Errorf("could not parse ogg sample rate")
	}

	tailSize := int64(128 * 1024)
	if fileSize < tailSize {
		tailSize = fileSize
	}
	tail := make([]byte, tailSize)
	if _, err := r.Seek(-tailSize, io.SeekEnd); err != nil {
		return 0, err
	}
	if _, err := io.ReadFull(r, tail); err != nil {
		return 0, err
	}
	granule := maxOggGranule(tail)
	if granule == 0 {
		return 0, fmt.Errorf("could not parse ogg granule")
	}
	return float64(granule) / sampleRate, nil
}

func parseOggSampleRateFromStart(data []byte) float64 {
	for i := 0; i+27 < len(data); i++ {
		if data[i] != 'O' || data[i+1] != 'g' || data[i+2] != 'g' || data[i+3] != 'S' {
			continue
		}
		if data[i+5] != 0x02 {
			continue
		}
		if rate := parseVorbisSampleRate(data[i+27:]); rate > 0 {
			return rate
		}
	}
	return 0
}

func maxOggGranule(data []byte) uint64 {
	var granule uint64
	for i := 0; i+14 < len(data); i++ {
		if data[i] != 'O' || data[i+1] != 'g' || data[i+2] != 'g' || data[i+3] != 'S' {
			continue
		}
		if data[i+5]&0x04 != 0 {
			continue
		}
		g := binary.LittleEndian.Uint64(data[i+6 : i+14])
		if g > granule {
			granule = g
		}
	}
	return granule
}

func parseVorbisSampleRate(packet []byte) float64 {
	needle := []byte{0x01, 'v', 'o', 'r', 'b', 'i', 's'}
	idx := -1
	for i := 0; i+len(needle) < len(packet); i++ {
		match := true
		for j := range needle {
			if packet[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			idx = i
			break
		}
	}
	if idx < 0 || idx+15+8 > len(packet) {
		return 0
	}
	pos := idx + 15
	sampleRateBits := binary.LittleEndian.Uint32(packet[pos : pos+4])
	return float64(sampleRateBits)
}
