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

        pkgs.keyutils
      ];
    };

  testScript = ''
    machine.wait_for_unit("default.target")

    # sanity check
    machine.succeed("ls /run")

    ### FS - LIST DIR ###
    ${fail "--rx /nix -- ls /run"}
    ${fail "--rx /nix --ro /etc -- ls /run"}

    ${succeed "--rx /nix --ro /run -- ls /run"}
    ${succeed "--rx /nix --rx /run -- ls /run"}
    ${succeed "--rx /nix --rw /run -- ls /run"}

    ${succeed "--unrestricted-fs -- ls /run"}

    ### FS - READ FILE ###
    ${fail "--rx /nix -- cat /etc/machine-id"}
    ${fail "--rx /nix --ro /etc/alsa -- cat /etc/machine-id"}

    ${succeed "--rx /nix --ro /etc -- cat /etc/machine-id"}
    ${succeed "--rx /nix --ro /etc/machine-id -- cat /etc/machine-id"}
    ${succeed "--unrestricted-fs -- cat /etc/machine-id"}

    ### FS - MAKE DIR ###
    ${fail "--rx /nix -- mkdir /tmp/dir1"}
    ${fail "--rx /nix --ro /tmp -- mkdir /tmp/dir2"}
    ${fail "--rx /nix --rx /tmp -- mkdir /tmp/dir3"}

    ${succeed "--rx /nix --rw /tmp -- mkdir /tmp/dir4"}
    ${succeed "--unrestricted-fs -- mkdir /tmp/dir5"}

    ### FS - MAKE FILE ###
    ${fail "--rx /nix -- touch /tmp/file1"}
    ${fail "--rx /nix --ro /tmp -- touch /tmp/file2"}
    ${fail "--rx /nix --rx /tmp -- touch /tmp/file3"}

    ${succeed "--rx /nix --rw /tmp -- touch /tmp/file4"}

    ### FS - REMOVE DIR ###
    machine.succeed("mkdir -p /tmp/remove-dir-test1 /tmp/remove-dir-test2")

    ${fail "--rx /nix -- rmdir /tmp/remove-dir-test1"}
    ${fail "--rx / -- rmdir /tmp/remove-dir-test1"}
    ${fail "--rx / --ro /tmp -- rmdir /tmp/remove-dir-test1"}
    ${fail "--rx / --rw /etc -- rmdir /tmp/remove-dir-test1"}

    # "fail" since according to kernel docs LANDLOCK_ACCESS_FS_REMOVE_DIR only allows
    # deleting dirs below the target dir, not the target dir itself
    ${fail "--rx / --rw /tmp/remove-dir-test1 -- rmdir /tmp/remove-dir-test1"}

    ${succeed "--rx / --rw /tmp -- rmdir /tmp/remove-dir-test1"}
    ${succeed "--unrestricted-fs -- rmdir /tmp/remove-dir-test2"}

    ### FS - WRITE FILE ###
    machine.succeed("touch /tmp/write-test")

    ${fail "--rx /nix -- bash -c 'echo 1 >> /tmp/write-test'"}
    ${fail "--rx /nix --ro /tmp -- bash -c 'echo 2 >> /tmp/write-test'"}
    ${fail "--rx / -- bash -c 'echo 3 >> /tmp/write-test'"}
    ${fail "--rx /nix --rw /etc -- bash -c 'echo 4 >> /tmp/write-test'"}

    ${succeed "--rx /nix --rw /tmp/write-test -- bash -c 'echo 5 >> /tmp/write-test'"}
    ${succeed "--rx /nix --rw /tmp -- bash -c 'echo 6 >> /tmp/write-test'"}
    ${succeed "--rx /nix --rw / -- bash -c 'echo 7 >> /tmp/write-test'"}

    ${succeed "--unrestricted-fs -- bash -c 'echo 8 >> /tmp/write-test'"}

    ### FS - EXECUTE ###
    ${fail "-- pwd"}
    ${fail "--ro / -- pwd"}
    ${fail "--rw / -- pwd"}
    ${fail "--rx /etc -- pwd"}

    ${succeed "--rx / -- pwd"}
    ${succeed "--unrestricted-fs -- pwd"}

    ### NET - TCP BIND ###
    ${tcpBindTest "fail" ""}
    ${tcpBindTest "fail" "--af inet"}
    ${tcpBindTest "fail" "--no-seccomp"}
    ${tcpBindTest "fail" "--bind-tcp 8000"}
    ${tcpBindTest "fail" "--af inet --connect-tcp 8000"}

    ${tcpBindTest "succeed" "--af inet --bind-tcp 8000"}
    ${tcpBindTest "succeed" "--no-seccomp --bind-tcp 8000"}
    ${tcpBindTest "succeed" "--unrestricted-net"}

    ### IPC - SIGNALS ###
    machine.fail("${./signal.sh}")
    machine.succeed("${./signal.sh} unscoped")

    ### SECCOMP - UNIX SOCKETS ###
    ${fail "--rx /nix -- busctl"}
    ${fail "--rx /nix --af inet -- busctl"}

    ${succeed "--rx /nix --af unix -- busctl"}
    ${succeed "--rx /nix --no-seccomp -- busctl"}

    ### SECCOMP - KEYRING SYSCALLS ###

    # sanity check
    machine.succeed("keyctl list @us")

    ${fail "--unrestricted-fs -- keyctl list @us"}
    ${fail "--unrestricted-fs --syscalls chmod -- keyctl list @us"}

    ${succeed "--unrestricted-fs --syscalls keyring -- keyctl list @us"}
    ${succeed "--unrestricted-fs --no-seccomp -- keyctl list @us"}

    ### SECCOMP - CHMOD SYSCALLS ###
    machine.succeed("mkdir -p /tmp/chmod-test")

    ${fail "--rx / -- chmod +r /tmp/chmod-test"}
    ${fail "--rx / --rw /tmp -- chmod +r /tmp/chmod-test"}
    ${fail "--rx / --syscalls keyring -- chmod +r /tmp/chmod-test"}

    ${succeed "--rx / --syscalls chmod -- chmod +r /tmp/chmod-test"}
    ${succeed "--rx / --no-seccomp -- chmod +r /tmp/chmod-test"}
    ${succeed "--unrestricted-fs -- chmod +r /tmp/chmod-test"}

    ### SECCOMP - CHOWN SYSCALLS ###
    machine.succeed("mkdir -p /tmp/chown-test")

    ${fail "--rx / -- chown root /tmp/chown-test"}
    ${fail "--rx / --rw /tmp -- chown root /tmp/chown-test"}
    ${fail "--rx / --syscalls keyring -- chown root /tmp/chown-test"}

    ${succeed "--rx / --syscalls chown -- chown root /tmp/chown-test"}
    ${succeed "--rx / --no-seccomp -- chown root /tmp/chown-test"}
    ${succeed "--unrestricted-fs -- chown root /tmp/chown-test"}

    ### ARG PARSING ###
    ${succeed "--rx /nix --rx /etc --rx /run -- ls /etc /run"}

    ${succeed "--rx /nix,/etc,/run -- ls /etc /run"}
    ${succeed "--rx=/nix,/etc,/run -- ls /etc /run"}

    ${succeed "--rx /nix,/etc --rx=/run -- ls /etc /run"}
  '';
}
