package musicsearch

import (
	"strings"
	"testing"
)

func TestUnwrapJSONPayload(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain json",
			in:   `{"code":0}`,
			want: `{"code":0}`,
		},
		{
			name: "callback jsonp",
			in:   `callback({"code":0})`,
			want: `{"code":0}`,
		},
		{
			name: "qq jsonp",
			in:   `MusicJsonCallback({"code":0,"data":{}})`,
			want: `{"code":0,"data":{}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := unwrapJSONPayload(tc.in)
			if got != tc.want {
				t.Fatalf("unwrapJSONPayload() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeImageURL(t *testing.T) {
	if got := normalizeImageURL("//p1.music.126.net/x.jpg"); got != "https://p1.music.126.net/x.jpg" {
		t.Fatalf("protocol-relative URL = %q", got)
	}
	if got := normalizeImageURL("http://p1.music.126.net/x.jpg"); !strings.HasPrefix(got, "https://") {
		t.Fatalf("http URL = %q", got)
	}
}

func TestKuwoPic(t *testing.T) {
	song := map[string]interface{}{
		"MVPIC": "/76/65/3389188905.jpg",
	}
	got := kuwoPic(song, "https://img3.kuwo.cn/star/albumcover/300")
	want := "https://img3.kuwo.cn/star/albumcover/300/76/65/3389188905.jpg"
	if got != want {
		t.Fatalf("kuwoPic() = %q, want %q", got, want)
	}
}

func TestFormatInterval(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{name: "milliseconds", value: 270000, want: "4:30"},
		{name: "seconds", value: 270, want: "4:30"},
		{name: "text mm:ss", value: "04:30", want: "4:30"},
		{name: "text m:ss", value: "4:30", want: "4:30"},
		{name: "text seconds", value: "270", want: "4:30"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatInterval(tc.value); got != tc.want {
				t.Fatalf("formatInterval() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestQQPic(t *testing.T) {
	song := map[string]interface{}{"albummid": "000v14Zi196WkA"}
	got := qqPic(song)
	if !strings.Contains(got, "000v14Zi196WkA") {
		t.Fatalf("qqPic() = %q", got)
	}
}
