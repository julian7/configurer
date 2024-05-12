package configurer

// Configuration holds applications' configuration.
//
// It can be anything, as long as diff3 can handle it
type Configuration any

// ConfigLoader can load configuration on demand.
//
// Notifier relies on the fact the configuration file is coming from a file in
// the operating system, and it uses fsnotify package to watch the file system.
type ConfigLoader interface {
	// Filename returns the file name which can then be watched by fsnotify.
	Filename() string
	// Load provides a new Configuration from the filename on demand. Otherwise,
	// it should produce an error.
	Load() (Configuration, error)
}
