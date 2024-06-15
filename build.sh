cd "$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1 && pwd)" || exit 1  # change to project directory
declare -A build_targets
build_targets=(["amd64"]="linux windows darwin" ["arm64"]="linux android darwin")
rule_filename="rules.dat"
dist_name=${1}

function build () {
    go build -o "./${1}${2}" "${3}"
    zip "${1}.zip" "./${1}" ${rule_filename} config.json
    sha256sum "${1}.zip" > "${1}.zip.sha256sum"
}

if [ -e ${rule_filename} ]
then
  rm ${rule_filename}
fi
wget https://github.com/HerrKKK/domain-list-community/releases/latest/download/${rule_filename}
for key in ${!build_targets[*]}
  do
    GOARCH=${key}
    go_os_array=${build_targets[$key]}
    for os in ${go_os_array[@]}
    do
      GOOS=${os}
      filename="${GOOS}-${GOARCH}-${dist_name}"
      if [ "${os}" == "windows" ]
      then
        build "./${filename}" "_cli.exe" "-ldflags=\"-H windowsgui\""
        build "./${filename}" ".exe"
      fi
      build "./${GOOS}-${GOARCH}-${dist_name}"
    done
  done
