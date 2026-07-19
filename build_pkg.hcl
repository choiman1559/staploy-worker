build "staploy-worker" {
  output_dir  = "out"
  version     = "shell:printf '%s' $APP_VERSION"
  executable = ["staploy"]
  lib_version = "shell:/home/$USER/sdk/go1.26.2/bin/go version | tr -d '\\n'"

  i386 { path = "out/386" }
  x86_64 { path = "out/amd64" }
  arm { path = "out/arm" }
  aarch64 { path = "out/arm64" }
  riscv64 { path = "out/riscv64" }
  mipsel { path = "out/mipsle" }
  mips64el { path = "out/mips64le" }
}