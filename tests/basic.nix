{ pkgs, lib, ... }:

let

  succeed = args: ''machine.succeed("icelock ${args}")'';
  fail = args: ''machine.fail("icelock ${args}")'';

  tcpBindTest =
    type: icelockArg:
    # SIGINT so that python exits with 0
    ''machine.${type}("timeout --signal=SIGINT --preserve-status 3s icelock --rx / ${icelockArg} -- ${lib.getExe pkgs.python3} -m http.server")'';

in

pkgs.testers.runNixOSTest {
  name = "basic";
  nodes.machine =
    { config, pkgs, ... }:
    {
      environment.systemPackages = [
        (pkgs.callPackage ../package.nix { })
      ];
    };

  testScript = ''
    machine.wait_for_unit("default.target")

    # sanity check
    machine.succeed("ls /run")

    ${fail "-- ls /run"}
    ${fail "--rx /nix/store -- ls /run"}
    ${succeed "--rx /nix/store --ro /run -- ls /run"}

    ${succeed "--rx / -- ls /run"}
    ${fail "--ro / -- ls /run"}
    ${fail "--rw / -- ls /run"}

    ${succeed "--unrestricted-fs -- ls /run"}

    ${fail "--rx /nix/store -- touch /tmp/something"}
    ${fail "--rx /nix/store --ro /tmp -- touch /tmp/something"}
    ${fail "--rx /nix/store --rx /tmp -- touch /tmp/something"}
    ${succeed "--rx /nix/store --rw /tmp -- touch /tmp/something"}

    machine.fail("${./signal.sh}")
    machine.succeed("${./signal.sh} unscoped")

    ${tcpBindTest "fail" ""}
    ${tcpBindTest "fail" "--af inet"}
    ${tcpBindTest "fail" "--no-seccomp"}
    ${tcpBindTest "fail" "--bind-tcp 8000"}
    ${tcpBindTest "fail" "--af inet --connect-tcp 8000"}

    ${tcpBindTest "succeed" "--af inet --bind-tcp 8000"}
    ${tcpBindTest "succeed" "--no-seccomp --bind-tcp 8000"}
    ${tcpBindTest "succeed" "--unrestricted-net"}

    ${fail "--unrestricted-fs -- busctl"}
    ${succeed "--unrestricted-fs --af unix -- busctl"}
    ${fail "--unrestricted-fs --af inet -- busctl"}
    ${succeed "--unrestricted-fs --no-seccomp -- busctl"}
  '';
}
