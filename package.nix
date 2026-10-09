{
  lib,
  buildGoModule,
  pkg-config,
  libseccomp,
  ...
}:

buildGoModule {
  name = "icelock";
  src = ./src;

  vendorHash = "sha256-r78I1T49dWUjP0ysc9WUVORZkneyMlGz44JMnTzTDPY=";

  nativeBuildInputs = [
    pkg-config
  ];

  buildInputs = [
    libseccomp
  ];

  meta = {
    description = "Tool for restricting programs with Landlock (and seccomp)";
    homepage = "https://github.com/2kool4idkwhat/icelock";
    license = lib.licenses.mit;
    platforms = lib.platforms.linux;
    mainProgram = "icelock";
  };
}
