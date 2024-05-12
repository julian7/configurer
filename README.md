# Configurer

Configuration handling for go server applications

This package maintains configuration file reads and updates across a long-running go application.

## Usage

Implement the following interfaces to your liking:

```go
type Config struct {
	ThisEntry string `json:"this_entry"`
	ThatEntry int    `json:"that_entry"`
}

func defaultConfig() *Config {
	return &Config{
		ThisEntry: "foo",
		ThatEntry: "baz",
	}
}

type Loader struct {
	filename string
}

func NewLoader(string filename) *Loader {
	return &Loader{filename: filename}
}

func (loader *Loader) Filename() string {
	return loader.filename
}

func (loader *Loader) Load() (*Config, error) {
	contents, err := os.ReadFile(loader.filename)
	if err != nil {
		return nil, fmt.Erorf("reading config file: %w", err)
	}

	conf := defaultConfig()
	err = json.Unmarshal(contents, conf)
	if err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return conf, nil
}
```

To read config:

```go
func runServer() error {
	ctx := context.Background()

	loader, err := NewLoader("config.json")
	if err != nil {
		return err
	}

	cctrl, err := configurer.New(loader, slog.Default())
	if err != nil {
		return err
	}
	...
}
```

Then, configuration can be read from `cctrl.Config()` directly. However, portions of confuguration handling can be spread across different subsystems via Notifier:

```go
func runServer() error {
	...
	// set up services
	someService := NewSomeservice(ctx)
	otherService := NewOtherservice(ctx)

	notifier := configurer.NewNotify(ctx, cctrl, slog.Default())

	notifier.RegisterServices(someService, otherService)
	notifier.RegisterAborters(someService)

	// sends initial configuration notifications to services
	if err := notifier.Notify(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if watch {
		// starts file watcher on configuration file
		notifier.Watch()
	}

	return someService.Run(otherService)
}
```

To consume config at services:

```go
...
func (svc *someService) UpdateConfig(ctx context.Context, ctrl *configurer.Control) error {
	cfg, ok := ctrl.Config().(*Config)
	if !ok {
		return errors.New("unknown configuration type")
	}

	if ctrl.IsChanged("ThisEntry") {
		...
	}

	return nil
}
```

To implement aborter:

```go
func (svc *someService) Abort(err error) {
	svc.logger.Warn("someService initiates shutdown due to error", "error", err)
	svc.Server.Shutdown()
}
```

## Licensing

SPDX-License-Identifier: BlueOak-1.0.0 OR MIT

This software is licensed under two licenses of your choice: Blue Oak Public License 1.0, or MIT Public License.
