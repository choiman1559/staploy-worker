alias {
  app "staploy-build" {
    name    = "staploy-worker"
    version = "shell:printf '%s' $APP_VERSION"
  }
}


build "alias:staploy-build" {
  output_dir  = "out"
  executable = ["staploy"]
  lib_version = "shell:go version"

  i386 { path = "out/386" }
  x86_64 { path = "out/amd64" }
  arm { path = "out/arm" }
  aarch64 { path = "out/arm64" }
  riscv64 { path = "out/riscv64" }
  mipsel { path = "out/mipsle" }
  mips64el { path = "out/mips64le" }
  mips { path = "out/mips" }
  mips64 { path = "out/mips64" }
  ppc64le { path = "out/ppc64le" }
  s390x { path = "out/s390x" }
}

configure {
  address      = "shell:echo $STAPLOY_HOST_ADDR"
  port         = "shell:echo $STAPLOY_HOST_PORT"
  enforce_uuid = false
}

manage "alias:staploy-build" {
  upload {}
}
