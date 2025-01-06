# TODO

- [ ] 01.03 Use fsnotify to reload TLS certificates when they are detected as changed
- [X] 02.04.24 Use fsnotify to update articles rather than interval based scanning
- [~] ~01.30.23 Modify internal/auth.Claims to contain a map of strings to values for claims rather than just a username value.~
- [~] ~01.30.23 Distinguish "pure" packages like internal/auth or internal/errors which don't interact with any bastion domain types from packages like internal/content or internal/router which expect to receive and return bastion domain types. I'm currently thinking internal/auth and internal/errors should be in a different directory from internal/router and internal/content, though I'm not sure of the name yet.~
