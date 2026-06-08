package local

import (
	"encoding/binary"
	"testing"
)

func TestParseTagDurationSeconds(t *testing.T) {
	if got := parseTagDurationSeconds("245000"); got != 245 {
		t.Fatalf("expected 245s from TLEN ms, got %v", got)
	}
	if got := parseTagDurationSeconds(""); got != 0 {
		t.Fatalf("expected 0 for empty string, got %v", got)
	}
}

func TestID3v2SyncSafeSize(t *testing.T) {
	size := id3v2SyncSafeSize([]byte{0x00, 0x00, 0x10, 0x00})
	if size != 2048 {
		t.Fatalf("expected 2048, got %d", size)
	}
}

func TestXingDurationSeconds(t *testing.T) {
	frame := make([]byte, 128)
	copy(frame[20:24], []byte("Xing"))
	binary.BigEndian.PutUint32(frame[24:28], 0x01) // frames flag
	binary.BigEndian.PutUint32(frame[28:32], 100)  // 100 frames

	seconds, ok := xingDurationSeconds(frame, 44100, 1152)
	if !ok {
		t.Fatal("expected xing duration")
	}
	want := float64(100*1152) / 44100
	if seconds < want-0.01 || seconds > want+0.01 {
		t.Fatalf("expected ~%v, got %v", want, seconds)
	}
}

func TestFormatTrackDuration(t *testing.T) {
	if got := formatTrackDuration(245.4); got != "4:05" {
		t.Fatalf("expected 4:05, got %q", got)
	}
	if got := formatTrackDuration(0); got != "" {
		t.Fatalf("expected empty duration, got %q", got)
	}
}
