{
  lib,
  buildGoModule,
  installShellFiles,
  pkg-config,
  libseccomp,
  ...
}:

buildGoModule {
  name = "icelock";
  src = ./src;

  vendorHash = "sha256-gRlxUIcBvkpdYUD7k1M/zsxtno6t/c3N7ly+cH9yi6s=";

  nativeBuildInputs = [
    installShellFiles

    pkg-config
  ];

  buildInputs = [
    libseccomp
  ];

  postInstall = ''
    installShellCompletion --cmd icelock \
      --bash <($out/bin/icelock completion bash) \
      --fish <($out/bin/icelock completion fish) \
      --zsh <($out/bin/icelock completion zsh)
  '';

  meta = {
    description = "Tool for restricting programs with Landlock";
    homepage = "https://github.com/2kool4idkwhat/icelock";
    license = lib.licenses.mit;
    platforms = lib.platforms.linux;
    mainProgram = "icelock";
  };
}
