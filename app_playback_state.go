package main

import (
	"sync"
	"time"

	"github.com/sunnyegg/miru/internal/storage"
)

const playbackPersistInterval = 15 * time.Second

type playbackStateUpdate struct {
	anilistID     int
	episodeNumber int
	position      float64
	percent       float64
}

type playbackStateWriter struct {
	store         *storage.Store
	pendingMu     sync.Mutex
	pending       *playbackStateUpdate
	flushRequests chan chan error
	stop          chan struct{}
	done          chan struct{}
	closeOnce     sync.Once
	closeErrMu    sync.Mutex
	closeErr      error
}

func newPlaybackStateWriter(store *storage.Store) *playbackStateWriter {
	writer := &playbackStateWriter{
		store:         store,
		flushRequests: make(chan chan error),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}
	go writer.run()
	return writer
}

func (w *playbackStateWriter) run() {
	ticker := time.NewTicker(playbackPersistInterval)
	defer ticker.Stop()
	defer close(w.done)

	for {
		select {
		case <-ticker.C:
			w.recordError(w.flushPending())
		case response := <-w.flushRequests:
			response <- w.flushPending()
		case <-w.stop:
			w.recordError(w.flushPending())
			return
		}
	}
}

func (w *playbackStateWriter) Submit(update playbackStateUpdate) {
	if update.anilistID <= 0 || update.episodeNumber <= 0 {
		return
	}
	if update.position < 0 {
		update.position = 0
	}

	w.pendingMu.Lock()
	w.pending = &update
	w.pendingMu.Unlock()
}

func (w *playbackStateWriter) Flush() error {
	response := make(chan error, 1)
	select {
	case <-w.done:
		return w.getCloseError()
	case w.flushRequests <- response:
		return <-response
	}
}

func (w *playbackStateWriter) Close() error {
	w.closeOnce.Do(func() {
		close(w.stop)
		<-w.done
	})
	return w.getCloseError()
}

func (w *playbackStateWriter) flushPending() error {
	w.pendingMu.Lock()
	update := w.pending
	w.pending = nil
	w.pendingMu.Unlock()
	if update == nil {
		return nil
	}
	return w.store.UpsertPlaybackState(
		update.anilistID,
		update.episodeNumber,
		update.position,
		update.percent,
	)
}

func (w *playbackStateWriter) recordError(err error) {
	if err == nil {
		return
	}
	w.closeErrMu.Lock()
	if w.closeErr == nil {
		w.closeErr = err
	}
	w.closeErrMu.Unlock()
}

func (w *playbackStateWriter) getCloseError() error {
	w.closeErrMu.Lock()
	defer w.closeErrMu.Unlock()
	return w.closeErr
}
