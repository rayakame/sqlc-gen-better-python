package log_test

import (
	"errors"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/log"
)

// TestL exercises the process-global singleton sequentially in one test:
// the empty-buffer Export must observe the buffer before anything logs.
func TestL(t *testing.T) {
	t.Parallel()

	name, data := log.L().Export()
	if name != "log.json" {
		t.Errorf("Export() name = %q, want %q", name, "log.json")
	}

	if string(data) != "[\n]" {
		t.Errorf("Export() data = %q, want %q", string(data), "[\n]")
	}

	first, second := log.L(), log.L()
	if first != second {
		t.Error("L() returned different instances across calls")
	}

	log.L().Log("first")
	log.L().Log("second")

	want := "[\n{\"message\":\"first\"},\n{\"message\":\"second\"}\n]"
	if _, got := log.L().Export(); string(got) != want {
		t.Errorf("Export() data = %q, want %q", string(got), want)
	}
}

func TestLoggerLog(t *testing.T) {
	t.Parallel()

	logger := &log.Logger{}
	logger.Log("hello world")

	want := "[\n{\"message\":\"hello world\"}\n]"
	if _, data := logger.Export(); string(data) != want {
		t.Errorf("Export() data = %q, want %q", string(data), want)
	}
}

func TestLoggerLogErr(t *testing.T) {
	t.Parallel()

	logger := &log.Logger{}
	logger.LogErr("reading file", errors.New("boom"))

	want := "[\n{\"error\":\"reading file: boom\"}\n]"
	if _, data := logger.Export(); string(data) != want {
		t.Errorf("Export() data = %q, want %q", string(data), want)
	}
}

func TestLoggerLogAny(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		message any
		want    string
	}{
		{
			name: "marshalable struct",
			message: struct {
				Key string `json:"key"`
			}{Key: "value"},
			want: "[\n{\"key\":\"value\"}\n]",
		},
		{
			name:    "unmarshalable channel hits marshal error branch",
			message: make(chan int),
			want:    "[\n{\"error\": \"Error while trying to log any: json: unsupported type: chan int\"}\n]",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			logger := &log.Logger{}
			logger.LogAny(tc.message)
			if _, data := logger.Export(); string(data) != tc.want {
				t.Errorf("Export() data = %q, want %q", string(data), tc.want)
			}
		})
	}
}

func TestLoggerExport(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		messages []string
		want     string
	}{
		{name: "empty buffer", want: "[\n]"},
		{name: "single message gets trailing newline", messages: []string{"one"}, want: "[\n{\"message\":\"one\"}\n]"},
		{
			name:     "messages joined with comma newline",
			messages: []string{"one", "two", "three"},
			want:     "[\n{\"message\":\"one\"},\n{\"message\":\"two\"},\n{\"message\":\"three\"}\n]",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			logger := &log.Logger{}
			for _, msg := range tc.messages {
				logger.Log(msg)
			}

			name, data := logger.Export()
			if name != "log.json" {
				t.Errorf("Export() name = %q, want %q", name, "log.json")
			}

			if string(data) != tc.want {
				t.Errorf("Export() data = %q, want %q", string(data), tc.want)
			}
		})
	}
}
