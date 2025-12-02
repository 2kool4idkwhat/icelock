{ pkgs, lib, ... }:

let

  tcpBindTest =
    type: icelockArg:
    # SIGINT so that python exits with 0
    ''machine.${type}("timeout --signal=SIGINT --preserve-status 3s icelock --rx / --af inet ${icelockArg} -- ${lib.getExe pkgs.python3} -m http.server")'';

in

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
    machine.fail("icelock --rw / -- ls /run")

    machine.succeed("icelock --unrestricted-fs -- ls /run")

    machine.fail("icelock --rx /nix/store -- touch /tmp/something")
    machine.fail("icelock --rx /nix/store --ro /tmp -- touch /tmp/something")
    machine.fail("icelock --rx /nix/store --rx /tmp -- touch /tmp/something")
    machine.succeed("icelock --rx /nix/store --rw /tmp -- touch /tmp/something")

    machine.fail("${./signal-scoped.sh}")
    machine.succeed("${./signal-unscoped.sh}")

    ${tcpBindTest "fail" ""}
    ${tcpBindTest "succeed" "--bind-tcp 8000"}
    ${tcpBindTest "fail" "--connect-tcp 8000"}
    ${tcpBindTest "succeed" "--unrestricted-net"}

    machine.fail("icelock --unrestricted-fs -- busctl")
    machine.succeed("icelock --unrestricted-fs --af unix -- busctl")
    machine.fail("icelock --unrestricted-fs --af inet -- busctl")
    machine.succeed("icelock --unrestricted-fs --no-seccomp -- busctl")
  '';
}
