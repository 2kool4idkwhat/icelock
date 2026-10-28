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

  vendorHash = "sha256-BnzDW0lmYSva2Ddc0Hwg/3AZ/cdU3nleEl+R2HJAizs=";

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
