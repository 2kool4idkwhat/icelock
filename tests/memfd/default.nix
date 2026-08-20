{
  lib,
  buildGoModule,
  ...
}:

buildGoModule {
  name = "memfd-test";
  src = ./.;

  vendorHash = "sha256-Np+MQ+oy8nyCBIT1ivJyt0sRpxgGkwGs8M9Je4oLt1I=";

  meta = {
    platforms = lib.platforms.linux;
    mainProgram = "memfd-test";
  };
}
