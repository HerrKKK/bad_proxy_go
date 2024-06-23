export ANDROID_NDK_HOME=${1}

export GOARCH=arm64
export GOOS=android
export CGO_ENABLED=1
export CC=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android31-clang

go build -buildmode=c-shared -o libproxy.so
echo "Build arm64-v8a success"
