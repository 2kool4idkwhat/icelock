{ pkgs, lib, ... }:
pkgs.testers.runNixOSTest {
  name = "landlock-disabled";
  nodes.machine =
    { config, pkgs, ... }:
    {
      environment.systemPackages = [
        (pkgs.callPackage ../package.nix { })
      ];

      security.lsm = lib.mkForce [
        # "landlock"
        "yama"
        "bpf"
      ];
    };

  testScript = ''
    machine.wait_for_unit("default.target")

    # sanity check
    machine.succeed("ls /run")

    machine.fail("icelock --unrestricted-fs -- ls /run")
  '';
}
