{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};

          makeIcelockWrapper = self.outputs.lib.${system}.makeIcelockWrapper;
        in
        {
          icelock = pkgs.callPackage ./package.nix { };
          default = self.outputs.packages.${system}.icelock;
        }
        // (import ./example.nix { inherit pkgs makeIcelockWrapper; })
      );

      lib = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        import ./lib.nix { inherit pkgs; }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              libseccomp
              pkg-config

              nil
            ];
          };
        }
      );

      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
          lib = pkgs.lib;
        in
        {
          basic = import ./tests/basic.nix { inherit pkgs lib; };
          landlock-disabled = import ./tests/landlock-disabled.nix { inherit pkgs lib; };
        }
      );

    };

}
