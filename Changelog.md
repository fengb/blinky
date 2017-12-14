# Blinky Changelog

## Unreleased

**FIXES**
* Multicast no longer starts to send without data
* systemd configured to restart on failure, workaround for multicast listening before network start

## v0.2.0

**FEATURES**
* Multicast communication between LAN servers
* Graceful reload of config via SIGHUP

**BREAKING**
* Rename conf section [web] to [http]
* Rename various configuration keys

## v0.1.1

**FIXES**
* Fixed memory leak from Pacman repo
* Better strategy to watch Pacman database changes
* Minor optimizations

## v0.1.0 - initial release

**FEATURES**
* Watch Pacman database changes
* HTTP server
* Auto refresh
