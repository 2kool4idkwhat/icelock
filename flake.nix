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

          wrap = self.outputs.lib.makeIcelockWrapper pkgs;
        in
        {
          icelock = pkgs.callPackage ./package.nix { };
          default = self.outputs.packages.${system}.icelock;

          mdwe-test = pkgs.callPackage ./tests/mdwe { };
          socket-test = pkgs.callPackage ./tests/sockets { };
        }
        // (import ./example.nix { inherit pkgs wrap; })
      );

      lib.makeIcelockWrapper =
        pkgs: module:
        (pkgs.lib.evalModules {
          modules = [
            ./module.nix
            module
          ];
          specialArgs = { inherit pkgs; };
        }).config.app.finalPackage;

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
