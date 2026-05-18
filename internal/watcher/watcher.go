package watcher

import (
	"context"
	"fmt"

	"github.com/fsnotify/fsnotify"
)

// Watch calls callback whenever path is written or created.
// Returns when ctx is canceled or a fatal error occurs.
func Watch(ctx context.Context, path string, callback func()) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer func() { _ = w.Close() }()

	if err := w.Add(path); err != nil {
		return fmt.Errorf("watch %q: %w", path, err)
	}

	return loop(ctx, w.Events, w.Errors, callback)
}

// loop processes fsnotify events until ctx is canceled, a channel closes, or a
// watcher error occurs.
func loop(
	ctx context.Context,
	events <-chan fsnotify.Event,
	errors <-chan error,
	callback func(),
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				callback()
			}
		case err, ok := <-errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}
