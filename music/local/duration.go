package local

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

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

func probeAudioDurationSeconds(absPath, ext string) float64 {
	switch strings.ToLower(ext) {
	case ".mp3":
		seconds, err := mp3DurationSeconds(absPath)
		if err == nil {
			return seconds
		}
	case ".flac":
		seconds, err := flacDurationSeconds(absPath)
		if err == nil {
			return seconds
		}
	case ".wav":
		seconds, err := wavDurationSeconds(absPath)
		if err == nil {
			return seconds
		}
	case ".ogg":
		seconds, err := oggDurationSeconds(absPath)
		if err == nil {
			return seconds
		}
	}
	return 0
}

func mp3DurationSeconds(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	decoder := mp3.NewDecoder(file)
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

func flacDurationSeconds(path string) (float64, error) {
	stream, err := flac.ParseFile(path)
	if err != nil {
		return 0, err
	}
	if stream.Info.SampleRate == 0 || stream.Info.NSamples == 0 {
		return 0, fmt.Errorf("invalid flac stream info")
	}
	return float64(stream.Info.NSamples) / float64(stream.Info.SampleRate), nil
}

func wavDurationSeconds(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var riff [12]byte
	if _, err := io.ReadFull(file, riff[:]); err != nil {
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
		if _, err := io.ReadFull(file, chunkHeader[:]); err != nil {
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
			if _, err := io.ReadFull(file, fmtChunk[:]); err != nil {
				return 0, err
			}
			byteRate = binary.LittleEndian.Uint32(fmtChunk[8:12])
			foundFmt = true
			if chunkSize > 16 {
				if _, err := file.Seek(int64(chunkSize-16), io.SeekCurrent); err != nil {
					return 0, err
				}
			}
		case "data":
			dataSize = chunkSize
			foundData = true
			if _, err := file.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return 0, err
			}
		default:
			if _, err := file.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return 0, err
			}
		}

		if chunkSize%2 == 1 {
			if _, err := file.Seek(1, io.SeekCurrent); err != nil {
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

func oggDurationSeconds(path string) (float64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	headSize := int64(96 * 1024)
	if stat.Size() < headSize {
		headSize = stat.Size()
	}
	head := make([]byte, headSize)
	if _, err := io.ReadFull(file, head); err != nil {
		return 0, err
	}
	sampleRate := parseOggSampleRateFromStart(head)
	if sampleRate <= 0 {
		return 0, fmt.Errorf("could not parse ogg sample rate")
	}

	tailSize := int64(128 * 1024)
	if stat.Size() < tailSize {
		tailSize = stat.Size()
	}
	tail := make([]byte, tailSize)
	if _, err := file.Seek(-tailSize, io.SeekEnd); err != nil {
		return 0, err
	}
	if _, err := io.ReadFull(file, tail); err != nil {
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
