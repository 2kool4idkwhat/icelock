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

  vendorHash = "sha256-KAQsQJ/5OaNMTWvhqekO80VD5o0VWfGxar7zT/Ic4Lg=";

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
    description = "Tool for restricting programs with Landlock (and seccomp)";
    homepage = "https://github.com/2kool4idkwhat/icelock";
    license = lib.licenses.mit;
    platforms = lib.platforms.linux;
    mainProgram = "icelock";
  };
}
