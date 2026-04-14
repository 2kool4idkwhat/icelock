{
  lib,
  buildGoModule,
  ...
}:

buildGoModule {
  name = "socket-test";
  src = ./.;

  vendorHash = "sha256-HT7jNscq9c/zu0po//Cc+lPV9qLcZOil1r20qZqVPFg=";

  meta = {
    platforms = lib.platforms.linux;
    mainProgram = "socket-test";
  };
}
