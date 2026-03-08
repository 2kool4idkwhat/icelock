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

  vendorHash = "sha256-Ts3rXoWLQ7hFSaoizP1tUh/NMqXQmIn0lczExe5MNoc=";

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
