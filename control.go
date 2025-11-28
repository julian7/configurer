package configurer

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/r3labs/diff"
)

// Control keeps track of the current and previous configurations, and a changelog.
type Control struct {
	changed  []string
	loader   ConfigLoader
	lock     sync.Mutex
	logger   *slog.Logger
	current  Configuration
	previous Configuration
}

// New returns a new Control object, or an error.
//
// It needs a ConfigLoader for reading the configuration, and a logger for logging
// purposes. A new configuration is read immediately.
func New(loader ConfigLoader, logger *slog.Logger) (*Control, error) {
	if loader == nil {
		return nil, ErrNoConfigFile
	}
	fileName := loader.Filename()
	if len(fileName) < 1 {
		return nil, ErrNoConfigFile
	}
	ctrl := &Control{
		loader: loader,
		logger: logger,
	}

	if err := ctrl.readConfig(); err != nil {
		return nil, err
	}

	return ctrl, nil
}

func (ctrl *Control) readConfig() error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	config, err := ctrl.loader.Load()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	ctrl.previous = ctrl.current
	ctrl.current = config

	if ctrl.previous == nil {
		ctrl.changed = []string{"*"}
		return nil
	}

	changelog, err := diff.Diff(ctrl.previous, ctrl.current)
	if err != nil {
		ctrl.logger.Warn("change diff unsuccessful", "error", err)
		return nil
	}

	if len(changelog) < 1 {
		ctrl.changed = []string{}
		return nil
	}

	ctrl.changed = make([]string, 0, len(changelog))
	for _, change := range changelog {
		ctrl.changed = append(ctrl.changed, strings.Join(change.Path, "."))
	}

	ctrl.logger.Debug("configuration changed", "changed", ctrl.changed)
	return nil
}

func (ctrl *Control) filename() string {
	return ctrl.loader.Filename()
}

// Config returns the current configuration. It needs to be casted to the final type.
func (ctrl *Control) Config() Configuration {
	return ctrl.current
}

// IsChanged confirms whether a certain portion of the struct has been changed.
//
// Item name is the configuration key's name. Deep configurations use "." as a
// delimiter. Numbers represent keys of slices.
//
// "*" can be used at the end of the string as a wildcard for any keys.
// For example, if "Database.Connection.Host" changes, the following wildcards
// will match: "*", "Database.*", "Database.Connection.*".
// Note: "*" matches full key, there's no substring matching.
func (ctrl *Control) IsChanged(item string) bool {
	op := func(entry string) bool { return item == entry }
	if snippy, ok := strings.CutSuffix(item, ".*"); ok {
		item = snippy + "."
		op = func(entry string) bool { return strings.HasPrefix(entry, item) }
	}
	for _, entry := range ctrl.changed {
		if entry == "*" {
			return true
		}
		if op(entry) {
			return true
		}
	}

	return false
}
