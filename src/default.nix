{
  buildGoModule,
  pkg-config,
  libseccomp,
  ...
}:

buildGoModule {
  name = "icelock";
  src = ./.;
  meta.mainProgram = "icelock";

  vendorHash = "sha256-gRlxUIcBvkpdYUD7k1M/zsxtno6t/c3N7ly+cH9yi6s=";

  nativeBuildInputs = [
    pkg-config
  ];

  buildInputs = [
    libseccomp
  ];
}
