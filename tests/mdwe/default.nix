{
  lib,
  buildGoModule,
  ...
}:

buildGoModule {
  name = "mdwe-test";
  src = ./.;

  vendorHash = "sha256-x+/YvOqc86xEf/x0DDE/1icLuw0KwSMQFaZEx6oGnTs=";

  meta = {
    platforms = lib.platforms.linux;
    mainProgram = "mdwe-test";
  };
}
