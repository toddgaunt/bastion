#! /bin/bash

# Run this script inside of the directory it resides in.
cd $(dirname $(realpath $0))
# Set much more strict behavior for failed commands and unexpanded variables.
set -eu
# Allow's the script to refer to itself.
me=$(basename $0)

function log() {
	echo "[$me] $@"
}

function increment-version() {
	local version=$(cat VERSION.txt)
	local old_version="$version"
	local a=( ${version//./ } )
	local major=${a[0]}
	local minor=${a[1]}
	local patch=${a[2]}

	while true; do
		read -p "Increment version [(m)ajor/m(i)nor/(p)atch]? " part
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

	local commit_msg="Version: $old_version -> $version"
	echo "$commit_msg"
	echo "$version" > VERSION.txt

	git add VERSION.txt
	git commit -m "$commit_msg"
	git tag "v$version"
}

function publish-version() {
			local version="v$(cat VERSION.txt)"
			git push origin "$version"
			GOPROXY="proxy.golang.org" go list -m "github.com/toddgaunt/persistent@$version"
}

function usage() {
	echo "Usage: $me [version|increment|publish|help]"
}

function main() {
	local subcommand=${1-version}
	local args=${@:2}

	case $subcommand in
		version)
			echo "$(cat VERSION.txt)"
			;;
		increment)
			increment-version $args
			;;
		publish)
			publish-version $args
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

main $@
