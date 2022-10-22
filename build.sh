#!/usr/bin/env bash

# Run this script inside of the directory it resides in.
cd $(dirname $(realpath $0))
# Set much more strict behavior for failed commands and unexpanded variables.
set -eu
# Allow's the script to refer to itself.
me=$(basename $0)

# Text colors
FMT_NONE="$(tput sgr0)"
FMT_RED="$(tput setaf 1)"
FMT_GREEN="$(tput setaf 2)"
FMT_YELLOW="$(tput setaf 3)"

package_files=("bastion" "docs" "www.example.com")
package=bastion.tar.gz

function log() {
	echo "[$me] $@"
}

function log-warn() {
	echo "[${FMT_YELLOW}${me}${FMT_NONE}] $@"
}

function log-fatal() {
	echo "[${FMT_RED}${me}${FMT_NONE}] $@"
	exit 1
}

# require() accepts a list of space separated commands and checks if they are
# available to be run. If all commands are available, returns with a value of
# 0. If any of the commands are unavailable, the script returns with a value of
# 1.
function require() {
	local missing=""
	for cmd in $@; do
		command -v "$cmd" >/dev/null 2>&1 || { missing="$missing $cmd"; }
	done

	# Even if only one pre-requisite is missing, exit since none are optional
	if [[ "$missing" != "" ]]; then
		log-fatal "please install required commands:$missing"
		exit 1
	fi
}

function build() {
	log "🔨 building…"
	go build
}

function clean() {
	log "🧹 cleaning…"
	go clean
	rm -f $package
}

function package() {
	log "📦 packaging…"
	if ! tar -cf - "$package_files" -P |
		pv -s $(du -sb "$package_files" | awk '{print $1}') |
		pigz > "$package"; then
		log-fatal "couldn't package $files"
	fi
}

function increment-version() {
	version=$(cat VERSION.txt)
	a=( ${version//./ } )
	major=${a[0]}
	minor=${a[1]}
	patch=${a[2]}

	while true; do
		read -p "Increment version [(M)ajor/m(I)nor/(P)atch]?" part
		if [[ "$part" == "" ]]; then
			part="M"
		fi
		case "$part" in
			[Mm])
				major=$((major + 1))
				minor=0
				patch=0
				break
			;;
			[Ii])
				minor=$((minor + 1))
				patch=0
				break
			;;
			[Pp])
				patch=$((patch + 1))
				break
			;;
			*)
				echo "Enter M to increment major version, I for minor version, or P for patch"
			;;
		esac
	done

	version="$major.$minor.$patch"
	echo "$version"
	echo "$version" > VERSION.txt
}

function usage() {
	echo "Usage: $me [all|clean|package|version|help]"
}

function main() {
	local subcommand=${1-all}
	local args=${@:2}

	case $subcommand in
		all)
			build $args
			;;
		clean)
			clean $args
			;;
		package)
			package $args
			;;
		version)
			increment-version $args
			;;
		-h|--help|help)
			usage
			;;
		*)
			log "$subcommand is not a valid subcommand."
			usage
			exit 1
			;;
	esac
}

require \
	tput \
	go \
	pigz \
	pv \
	realpath \
	tar

main $@
