# Bastion

[![Go](https://github.com/toddgaunt/bastion/actions/workflows/go.yml/badge.svg)](https://github.com/toddgaunt/bastion/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/toddgaunt/bastion.svg)](https://pkg.go.dev/github.com/toddgaunt/bastion)

Bastion is a simple webserver for serving articles written in markdown. These markdown
articles are converted to html automatically as they are updated, and have a special
header section to specify useful metadata about the article to Bastion.

## Article Authentication
Individual articles can require HTTP basic authentication if the article's source document includes values for the Username and Password in the article header.

For example:
```markdown
Title: My protected article
Username: monkey
Password: bananas
=== markdown ===
I want to protect the content of this article with HTTP basic auth so no one
can steal my secrets!
```

## Website layout
```
www.example.com/
    config.json
    content/
        about.md
        contact.md
    static/
        default.css
```

## Developer Quickstart
First, make sure the following programs are installed:
- go
- pigz
- pv
- realpath
- tar

Then run the following commands:
```
./build.sh build
./bastion www.example.com
```
For commandline options and usage information run `bastion -h`

## Versioning
Since bastion is mostly used for my own personal
[website](www.bastionburrow.com), it isn't going to be very stable. I plan on
changing things around on a whim. With this said, as long as bastion has a
major version of 0 (e.g. 0.1.12), minor versions will be treated as breaking
changes, and patch versions will be treated as backward-compatible changes.
Expect more for the former than the latter. If I one day decide to release
bastion, then I will start using semver properly.
