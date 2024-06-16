cd "$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit && pwd)" || exit  # change to project directory
declare -A build_targets
build_targets=(["amd64"]="linux windows darwin" ["arm64"]="linux android darwin")
CGO_ENABLE=0
dist_name=${1}

function build () {
    go build -o "./${1}" "${2}"
    zip "${1}.zip" "./${1}" rules.dat config.json
    sha256sum "${1}.zip" > "${1}.zip.sha256sum"
}

for key in ${!build_targets[*]}
  do
    GOARCH=${key}
    go_os_array=${build_targets[$key]}
    for os in "${go_os_array[@]}"
    do
      GOOS=${os}
      filename="${GOOS}-${GOARCH}-${dist_name}"
      if [ "${os}" == "windows" ]
      then
        build "./${filename}_cli.exe" "-ldflags=\"-H windowsgui\""
        build "./${GOOS}-${GOARCH}-${dist_name}.exe"
      fi
      build "./${GOOS}-${GOARCH}-${dist_name}"
    done
  done
