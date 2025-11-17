{
  buildGoModule,
  ...
}:

buildGoModule {
  name = "icelock";
  src = ./.;
  meta.mainProgram = "icelock";

  vendorHash = "sha256-s9NzEDxNtfUMmXbPqt/n2j/4ORN8MFaqdavmTRuNbqo=";
}
