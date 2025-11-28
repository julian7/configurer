# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

No changes so far.

## [v0.2.0] - Nov 28, 2025

Changed:

* Notifier.Watch: now it can be canceled by the provided context
* Notifier.RegisterServices: sends initial notification of configuration updates to services upon registration, if initial notification is already sent.
* Notifier.Notify: sends initial notification only once. Later calls to RegisterServices() will send initial notification to newly registered services individually.
* Don't fire update notification if updated config is identical to the previous one

## [v0.1.0] - May 12, 2024

Initial release.

[Unreleased]: https://github.com/julian7/configurer/
[v0.1.0]: https://github.com/julian7/configurer/tag/v0.1.0
[v0.2.0]: https://github.com/julian7/configurer/tag/v0.2.0
