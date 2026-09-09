package bearer

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantOK    bool
	}{
		{"valid header", "Bearer abc123", "abc123", true},
		{"missing header", "", "", false},
		{"wrong scheme", "Basic dXNlcjpwYXNz", "", false},
		{"prefix with no token", "Bearer ", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ok := Parse(tt.header)
			if token != tt.wantToken || ok != tt.wantOK {
				t.Errorf("Parse(%q) = (%q, %v), want (%q, %v)", tt.header, token, ok, tt.wantToken, tt.wantOK)
			}
		})
	}
}
