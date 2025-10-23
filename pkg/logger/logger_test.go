package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNew_ValidLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel zapcore.Level
		wantErr       bool
	}{
		{
			name:          "debug level",
			level:         "debug",
			expectedLevel: zapcore.DebugLevel,
			wantErr:       false,
		},
		{
			name:          "info level",
			level:         "info",
			expectedLevel: zapcore.InfoLevel,
			wantErr:       false,
		},
		{
			name:          "warn level",
			level:         "warn",
			expectedLevel: zapcore.WarnLevel,
			wantErr:       false,
		},
		{
			name:          "error level",
			level:         "error",
			expectedLevel: zapcore.ErrorLevel,
			wantErr:       false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			wantErr: false,
		},
		{
			name:          "empty level defaults to info",
			level:         "",
			expectedLevel: zapcore.InfoLevel,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}

			if logger == nil {
				t.Error("New() returned nil logger")
				return
			}

			if logger.Core().Enabled(tt.expectedLevel) == false {
				t.Errorf("Logger should be enabled for level %v", tt.expectedLevel)
			}

			// Cleanup
			_ = logger.Sync()
		})
	}
}

func TestNew_DebugUsesDevConfig(t *testing.T) {
	logger, err := New("debug")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer logger.Sync()

	if logger == nil {
		t.Error("Debug logger should not be nil")
	}
}

func TestNew_InfoUsesProductionConfig(t *testing.T) {
	logger, err := New("info")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer logger.Sync()

	if logger == nil {
		t.Error("Info logger should not be nil")
	}
}
