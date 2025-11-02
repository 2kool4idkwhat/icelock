{
  buildGoModule,
  ...
}:

buildGoModule {
  name = "icelock";
  src = ./.;
  meta.mainProgram = "icelock";

  vendorHash = "sha256-WWoSV50GJoJkA5InvgPJW7VzT3gGjlwJTFcOg+mENVU=";
}
