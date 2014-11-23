package youtube

import "testing"

func TestQuickMatch(t *testing.T) {
	tests := []struct {
		url   string
		match bool
	}{
		{"https://youtu.be/abcde-fg_12", true},
		{"http://youtu.be/abcde-fg_12", true},
		{"http://www.google.com/", false},
		{"https://www.google.com/", false},
		{"https://www.youtube.com/watch?v=zyx123wvu_4", true},
		{"https://www.youtube.com/watch?k=v&v=zyx123wvu_4", true},
		{"https://www.youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", true},
		{"https://www.youtube.com/watch?k=v&v=zyx123wvu!4&y=7&t=5", false},
		{"https://www.youtube.com/watch?v=zyx123wvu4", false},
		{"https://www.youtube.com/watch?v=zyx123wvu456", false},
		{"http://www.youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", true},
		{"http://youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", true},
		{"https://youtube.com/watch?k=v&v=zyx123wvu_4&y=7&t=5", true},
		{"https://www.youtube.com/v/youtubecode", true},
		{"https://www.youtube.com/v/youtubecode?y=1", true},
		{"http://www.youtube.com/v/youtubecode?y=1", true},
		{"https://youtube.com/v/youtubecode?y=1", true},
		{"http://youtube.com/v/youtubecode?y=1", true},
		{"http://youtube.com/v/outubecode?y=1", false},
		{"http://youtube.com/v/yyoutubecode?y=1", false},
	}

	for n, test := range tests {
		if quickMatch(test.url) != test.match {
			t.Errorf("test %d: expecting match %v, got %v", n+1, test.match, !test.match)
		}
	}
}
