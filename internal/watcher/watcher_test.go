package watcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatch(t *testing.T) {
	tests := []struct {
		name        string
		triggerFunc func(path string)
		wantTrigger bool
	}{
		{
			name: "callback called on file write",
			triggerFunc: func(path string) {
				_ = os.WriteFile(path, []byte("changed"), 0o644)
			},
			wantTrigger: true,
		},
		{
			name:        "no callback when file unchanged",
			triggerFunc: func(path string) {},
			wantTrigger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "mpkg.yaml")
			_ = os.WriteFile(path, []byte("initial"), 0o644)

			called := make(chan struct{}, 1)
			ctx := t.Context()

			go func() {
				_ = Watch(ctx, path, func() { called <- struct{}{} })
			}()

			time.Sleep(100 * time.Millisecond)
			tt.triggerFunc(path)

			select {
			case <-called:
				if !tt.wantTrigger {
					t.Error("callback triggered unexpectedly")
				}
			case <-time.After(800 * time.Millisecond):
				if tt.wantTrigger {
					t.Error("callback not triggered within timeout")
				}
			}
		})
	}
}

func TestWatchNonExistentPath(t *testing.T) {
	err := Watch(context.Background(), "/nonexistent/path/mpkg.yaml", func() {})
	if err == nil {
		t.Error("expected error watching non-existent path, got nil")
	}
}

func TestWatchContextCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mpkg.yaml")
	_ = os.WriteFile(path, []byte("initial"), 0o644)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, path, func() {})
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Watch() returned unexpected error after cancel: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Watch() did not return after context cancel")
	}
}

func TestLoop(t *testing.T) {
	tests := []struct {
		name     string
		run      func(events chan fsnotify.Event, errs chan error, cancel context.CancelFunc)
		wantCall bool
		wantErr  bool
	}{
		{
			name: "write event triggers callback",
			run: func(events chan fsnotify.Event, _ chan error, _ context.CancelFunc) {
				events <- fsnotify.Event{Op: fsnotify.Write, Name: "mpkg.yaml"}
			},
			wantCall: true,
		},
		{
			name: "create event triggers callback",
			run: func(events chan fsnotify.Event, _ chan error, _ context.CancelFunc) {
				events <- fsnotify.Event{Op: fsnotify.Create, Name: "mpkg.yaml"}
			},
			wantCall: true,
		},
		{
			name: "rename event does not trigger callback then channel closes",
			run: func(events chan fsnotify.Event, _ chan error, _ context.CancelFunc) {
				events <- fsnotify.Event{Op: fsnotify.Rename, Name: "mpkg.yaml"}
				close(events)
			},
		},
		{
			name: "events channel closed returns nil",
			run: func(events chan fsnotify.Event, _ chan error, _ context.CancelFunc) {
				close(events)
			},
		},
		{
			name: "errors channel sends error",
			run: func(_ chan fsnotify.Event, errs chan error, _ context.CancelFunc) {
				errs <- errors.New("disk full")
			},
			wantErr: true,
		},
		{
			name: "errors channel closed returns nil",
			run: func(_ chan fsnotify.Event, errs chan error, _ context.CancelFunc) {
				close(errs)
			},
		},
		{
			name: "context cancel returns nil",
			run: func(_ chan fsnotify.Event, _ chan error, cancel context.CancelFunc) {
				cancel()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := make(chan fsnotify.Event, 1)
			errs := make(chan error, 1)
			called := make(chan struct{}, 1)
			done := make(chan error, 1)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			tt.run(events, errs, cancel)

			go func() {
				done <- loop(ctx, events, errs, func() {
					called <- struct{}{}
					cancel()
				})
			}()

			select {
			case err := <-done:
				if (err != nil) != tt.wantErr {
					t.Errorf("loop() error = %v, wantErr %v", err, tt.wantErr)
				}
				select {
				case <-called:
					if !tt.wantCall {
						t.Error("callback triggered unexpectedly")
					}
				default:
					if tt.wantCall {
						t.Error("callback not triggered")
					}
				}
			case <-time.After(2 * time.Second):
				t.Error("loop() did not return within timeout")
			}
		})
	}
}
