package youtube

import "testing"

func TestGetCode(t *testing.T) {
	tests := []struct {
		url, code string
	}{
		{"https://youtu.be/abcde-fg_12", "abcde-fg_12"},
		{"http://youtu.be/abcde-fg_12", "abcde-fg_12"},
		{"http://www.google.com/", ""},
		{"https://www.google.com/", ""},
		{"https://www.youtube.com/watch?v=zyx123wvu_4", "zyx123wvu_4"},
		{"https://www.youtube.com/watch?k=v&v=zyx123wvu_4", "zyx123wvu_4"},
		{"https://www.youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", "zyx123wvu_4"},
		{"https://www.youtube.com/watch?k=v&v=zyx123wvu!4&y=7&t=5", ""},
		{"https://www.youtube.com/watch?v=zyx123wvu4", ""},
		{"https://www.youtube.com/watch?v=zyx123wvu456", ""},
		{"http://www.youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", "zyx123wvu_4"},
		{"http://youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", "zyx123wvu_4"},
		{"https://youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", "zyx123wvu_4"},
		{"https://www.youtube.com/v/youtubecode", "youtubecode"},
		{"https://www.youtube.com/v/youtubecode?y=1", "youtubecode"},
		{"http://www.youtube.com/v/youtubecode?y=1", "youtubecode"},
		{"https://youtube.com/v/youtubecode?y=1", "youtubecode"},
		{"http://youtube.com/v/youtubecode?y=1", "youtubecode"},
		{"http://youtube.com/v/outubecode?y=1", ""},
		{"http://youtube.com/v/yyoutubecode?y=1", ""},
	}

	for n, test := range tests {
		code := getCode(test.url)
		if code != test.code {
			t.Errorf("test %d: expecting code %q, got %q", n+1, test.code, code)
		}
	}
}
