package controller

import "testing"

func TestToolPortProcessNormalizePort(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    int
		wantErr bool
	}{
		{name: "string ok", input: "8080", want: 8080},
		{name: "int ok", input: 3000, want: 3000},
		{name: "zero invalid", input: 0, wantErr: true},
		{name: "too large invalid", input: 70000, wantErr: true},
		{name: "non number invalid", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeToolPort(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeToolPort(%v) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeToolPort(%v) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeToolPort(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestToolPortProcessParseWindowsNetstat(t *testing.T) {
	raw := "  TCP    0.0.0.0:8080           0.0.0.0:0              LISTENING       1234\r\n" +
		"  TCP    127.0.0.1:9000         0.0.0.0:0              LISTENING       4567\r\n" +
		"  TCP    [::]:8080              [::]:0                 LISTENING       1234\r\n"

	got := parseWindowsPortProcessRows(raw, 8080, map[int]string{
		1234: "node.exe",
	})

	if len(got) != 2 {
		t.Fatalf("row count = %d, want 2, got=%v", len(got), got)
	}
	if got[0].PID != 1234 {
		t.Fatalf("first pid = %d, want 1234", got[0].PID)
	}
	if got[0].Command != "node.exe" {
		t.Fatalf("first command = %q, want %q", got[0].Command, "node.exe")
	}
	if got[0].Protocol != "tcp" {
		t.Fatalf("first protocol = %q, want %q", got[0].Protocol, "tcp")
	}
	if got[0].Address != "0.0.0.0:8080" {
		t.Fatalf("first address = %q, want %q", got[0].Address, "0.0.0.0:8080")
	}
}

func TestToolPortProcessParseUnixLsof(t *testing.T) {
	raw := "COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME\n" +
		"node    12345 frog   22u  IPv4  12345      0t0  TCP *:8080 (LISTEN)\n" +
		"python3 23456 frog   11u  IPv6  54321      0t0  TCP 127.0.0.1:8080 (LISTEN)\n"

	got := parseUnixPortProcessRows(raw)

	if len(got) != 2 {
		t.Fatalf("row count = %d, want 2, got=%v", len(got), got)
	}
	if got[0].PID != 12345 {
		t.Fatalf("first pid = %d, want 12345", got[0].PID)
	}
	if got[0].Command != "node" {
		t.Fatalf("first command = %q, want %q", got[0].Command, "node")
	}
	if got[0].Protocol != "tcp" {
		t.Fatalf("first protocol = %q, want %q", got[0].Protocol, "tcp")
	}
	if got[0].Address != "*:8080" {
		t.Fatalf("first address = %q, want %q", got[0].Address, "*:8080")
	}
}
