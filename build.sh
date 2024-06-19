declare -A build_targets
build_targets=(["amd64"]="linux windows darwin" ["arm64"]="linux android darwin")
rule_filename="rules.dat"
dist_name=${1}

function prepare() {
    cd "$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1 && pwd)" || exit 1  # change to project directory
    if [ ! -d "./dist"  ]
    then
      mkdir ./dist
    fi
    wget https://github.com/HerrKKK/domain-list-community/releases/latest/download/${rule_filename}
}

function package_zip() {
    zip "${dist_name}-${GOOS}-${GOARCH}.zip" "./${1}" ${rule_filename} config.json
    sha256sum "${dist_name}-${GOOS}-${GOARCH}.zip" > "${dist_name}-${GOOS}-${GOARCH}.zip.sha256sum"
}

function package_tar_gz() {
    tar czvf "${dist_name}-${GOOS}-${GOARCH}.tar.gz" "./${1}" ${rule_filename} config.json
    sha256sum "${dist_name}-${GOOS}-${GOARCH}.tar.gz" > "${dist_name}-${GOOS}-${GOARCH}.tar.gz.sha256sum"
}


function main() {
    prepare
    for arch in ${!build_targets[*]}
      do
        go_os_array=${build_targets[$arch]}
        for os in ${go_os_array[@]}
        do
          export GOARCH=${arch}
          export GOOS=${os}
          filename="${dist_name}-${GOOS}-${GOARCH}"
          if [ "${GOOS}" == "windows" ]
          then
            go build -o "./${filename}.exe"
            package_zip "${filename}.exe"
            go build -o "./${filename}_nogui.exe" -ldflags="-H windowsgui"
            package_zip "${filename}_nogui.exe"
          else
            go build -o "./${filename}"
            package_tar_gz "${filename}"
          fi
        done
      done

    mv ./*.tar.gz ./dist
    mv ./*.zip ./dist
    mv ./*.sha256sum ./dist
}

main