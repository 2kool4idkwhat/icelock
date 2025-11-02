{ pkgs, ... }:
pkgs.testers.runNixOSTest {
  name = "basic";
  nodes.machine =
    { config, pkgs, ... }:
    {
      environment.systemPackages = [
        (pkgs.callPackage ../src { })
      ];
    };

  testScript = ''
    machine.wait_for_unit("default.target")

    # sanity check
    machine.succeed("ls /run")

    machine.fail("icelock -- ls /run")
    machine.fail("icelock --rx /nix/store -- ls /run")
    machine.succeed("icelock --rx /nix/store --ro /run -- ls /run")

    machine.succeed("icelock --rx / -- ls /run")
    machine.fail("icelock --ro / -- ls /run")

    machine.succeed("icelock --unrestricted-fs -- ls /run")
  '';
}
