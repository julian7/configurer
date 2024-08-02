package configurer

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Aborter provides abort notification to spread.
type Aborter interface {
	Abort(error)
}

// Updateable enables objects to have their configurations updated.
type Updateable interface {
	// UpdateConfig is called by Notifier on configuration change. When a change
	// cannot be handled for some reason, the error initiates abort.
	UpdateConfig(context.Context, *Control) error
}

// Notifier tells subsystems about configuration changes.
type Notifier struct {
	ctx      context.Context
	logger   *slog.Logger
	watcher  *fsnotify.Watcher
	services []Updateable
	aborters []Aborter
	ctrl     *Control
	wantdown bool
}

// NewNotifier returns a new notifier object.
func NewNotifier(ctx context.Context, ctrl *Control, logger *slog.Logger) *Notifier {
	return &Notifier{
		ctx:    ctx,
		ctrl:   ctrl,
		logger: logger,
	}
}

// RegisterServices adds Updateable services to the list of services to be notified.
func (notif *Notifier) RegisterServices(svc ...Updateable) {
	notif.services = append(notif.services, svc...)
}

// RegisterAborters adds Aborter services to the list of services handling abortions.
func (notif *Notifier) RegisterAborters(svc ...Aborter) {
	notif.aborters = append(notif.aborters, svc...)
}

// Notify sends configuration change notification to Updateable services.
//
// This method should be called right after services and aborters registered.
// It sends notification to the configuration object itself too, if it implements
// Updateable.
func (notif *Notifier) Notify() error {
	if cfg, ok := notif.ctrl.Config().(Updateable); ok {
		cfg.UpdateConfig(notif.ctx, notif.ctrl)
	}
	for _, svc := range notif.services {
		err := svc.UpdateConfig(notif.ctx, notif.ctrl)
		if err == nil {
			continue
		}
		if !notif.wantdown {
			notif.wantdown = true
			for _, svc := range notif.aborters {
				svc.Abort(err)
				return err
			}
		}
		return err
	}

	return nil
}

// Watch starts configuration file watching for changes using fsnotify.
//
// It handles modify and remove events. On removal, it tries to re-add
// the file to the watchlist immediately, and continues trying with an
// exponential backoff (starting with 1/2 seconds, with a multiplier of 1.5,
// backing off after 10 tries).
//
// Watch can be canceled by calling cancelFunc of the provided context.
func (notif *Notifier) Watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("setting up new watcher for config: %w", err)
	}

	notif.watcher = watcher
	if err = watcher.Add(notif.ctrl.filename()); err != nil {
		return fmt.Errorf("adding config file name to watcher: %w", err)
	}

	go notif.watch()

	return nil
}

func (notif *Notifier) watch() {
	defer func() {
		notif.logger.Info("watcher finished")
		notif.watcher.Close()
	}()

	for {
		select {
		case event, ok := <-notif.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				notif.modify(event)
			} else if event.Has(fsnotify.Remove) {
				notif.replace(event)
			}
		case err, ok := <-notif.watcher.Errors:
			if !ok {
				return
			}
			notif.logger.Warn("fsnotify error", "error", err)
		case <-notif.ctx.Done():
			return
		}
	}
}

func (notif *Notifier) modify(event fsnotify.Event) {
	notif.logger.Debug("configuration modified", "filename", event.Name)
	if err := notif.ctrl.readConfig(); err != nil {
		notif.logger.Warn("error reloading config", "error", err)
	}
	_ = notif.Notify()
}

func (notif *Notifier) replace(event fsnotify.Event) {
	notif.logger.Debug("configuration modified", "filename", event.Name)
	notif.readdWatcher(0)()

	if err := notif.ctrl.readConfig(); err != nil {
		notif.logger.Warn("error reloading config", "error", err)
	}
	_ = notif.Notify()
}

func (notif *Notifier) readdWatcher(attempt int) func() {
	baseDelay := 500 * time.Millisecond
	multiplier := 1.5
	delay := baseDelay * time.Duration(int64(math.Pow(float64(multiplier), float64(attempt))))

	return func() {
		if err := notif.watcher.Add(notif.ctrl.filename()); err != nil {
			if attempt > 10 {
				notif.logger.Warn("error re-adding config file watcher; disable watching", "error", err)
				return

			}
			notif.logger.Warn("error re-adding config file watcher", "error", err)
			time.AfterFunc(delay, notif.readdWatcher(attempt+1))
		}
	}
}
