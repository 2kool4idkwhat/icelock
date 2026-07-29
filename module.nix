{
  pkgs,
  lib,
  config,
  ...
}:
let
  inherit (lib) mkOption types;
  inherit (types)
    bool
    str
    int
    listOf
    ;

  icelockPkg = pkgs.callPackage ./package.nix { };
  icelock = lib.getExe icelockPkg;

  listOpt =
    flag: list: if list == [ ] then [ ] else ''${flag}="${builtins.concatStringsSep ''","'' list}"'';

  portListOpt = flag: list: listOpt flag (map builtins.toString list);

  boolOpt = flag: value: if value then flag else [ ];
in
{
  options = {
    app = {
      package = mkOption {
        type = types.package;
      };
      finalPackage = mkOption {
        type = types.package;
        readOnly = true;
      };

      extraBinPaths = mkOption {
        type = listOf str;
        default = [ ];
      };
      flags = mkOption {
        type = listOf str;
        default = [ ];
      };
      env = mkOption {
        default = { };
      };
    };

    fs = {
      restrict = mkOption {
        type = bool;
        default = true;
      };

      ro = mkOption {
        type = listOf str;
        default = [ ];
      };
      rx = mkOption {
        type = listOf str;
        default = [ "/nix/store" ];
      };
      rw = mkOption {
        type = listOf str;
        default = [ ];
      };
      unix = mkOption {
        type = listOf str;
        default = [ ];
      };
    };

    net = {
      restrict = mkOption {
        type = bool;
        default = true;
      };

      bind = mkOption {
        type = listOf int;
        default = [ ];
      };
      bindAll = mkOption {
        type = bool;
        default = false;
      };

      connect = mkOption {
        type = listOf int;
        default = [ ];
      };
      connectAll = mkOption {
        type = bool;
        default = false;
      };
    };

    signals = mkOption {
      type = bool;
      default = false;
    };
    abstractUnixSockets = mkOption {
      type = bool;
      default = false;
    };

    seccomp = {
      enable = mkOption {
        type = bool;
        default = true;
      };

      socketFamilies = mkOption {
        type = listOf str;
        default = [ ];
      };
      syscalls = mkOption {
        type = listOf str;
        default = [ ];
      };

      userNamespaces = mkOption {
        type = bool;
        default = false;
      };
      ioUring = mkOption {
        type = bool;
        default = false;
      };
      keyring = mkOption {
        type = bool;
        default = false;
      };
      posixMessageQueues = mkOption {
        type = bool;
        default = false;
      };
      sysvMessageQueues = mkOption {
        type = bool;
        default = false;
      };
    };

    mdwe = mkOption {
      type = bool;
      default = false;
    };
    keepCaps = mkOption {
      type = bool;
      default = false;
    };
  };

  config =
    let
      envVars = lib.mapAttrsToList (name: value: ''${name}='"${value}"' '') config.app.env;

      icelockArgs = builtins.concatStringsSep " " (
        lib.flatten [
          (boolOpt "--unrestricted-fs" (!config.fs.restrict))
          (listOpt "--ro" config.fs.ro)
          (listOpt "--rx" config.fs.rx)
          (listOpt "--rw" config.fs.rw)
          (listOpt "--unix" config.fs.unix)

          (boolOpt "--unrestricted-net" (!config.net.restrict))
          (portListOpt "--bind" config.net.bind)
          (boolOpt "--bind-all" config.net.bindAll)
          (portListOpt "--connect" config.net.connect)
          (boolOpt "--connect-all" config.net.connectAll)

          (boolOpt "--signals" config.signals)
          (boolOpt "--abstract-unix" config.abstractUnixSockets)

          (boolOpt "--no-seccomp" (!config.seccomp.enable))
          (listOpt "--af" config.seccomp.socketFamilies)
          (listOpt "--syscalls" config.seccomp.syscalls)
          (boolOpt "--userns" config.seccomp.userNamespaces)
          (boolOpt "--io-uring" config.seccomp.ioUring)
          (boolOpt "--keyring" config.seccomp.keyring)
          (boolOpt "--posix-mq" config.seccomp.posixMessageQueues)
          (boolOpt "--sysv-mq" config.seccomp.sysvMessageQueues)

          (boolOpt "--mdwe" config.mdwe)
          (boolOpt "--keep-caps" config.keepCaps)
        ]
      );
    in
    {
      app.finalPackage = pkgs.symlinkJoin {
        name = config.app.package.name;
        paths = [
          config.app.package
        ]
        ++ (if (builtins.hasAttr "man" config.app.package) then [ config.app.package ] else [ ]);
        passthru.unwrapped = config.app.package;
        postBuild = ''
          # TODO: make a bash function so we don't have duplicate wrapping commands for
          # implicit and explicit bins
          for file in "$out/bin/"*; do
            base=$(basename "$file")

            echo "wrapping $file"
            rm "$file"

            echo "#!${lib.getExe pkgs.bashNonInteractive}" > "$file"

            for var in ${builtins.concatStringsSep " " envVars}; do
              echo "export $var" >> "$file"
            done

            echo -n 'exec ${icelock} ${icelockArgs} ' >> "$file"
            echo "-- "${config.app.package}/bin/$base" ${builtins.concatStringsSep " " config.app.flags} \"\$@\"" >> "$file"

            chmod +x "$file"
          done

          for file in ${builtins.concatStringsSep " " config.app.extraBinPaths}; do
            path="$out/$file"

            echo "wrapping $path"
            rm "$path"

            echo "#!${lib.getExe pkgs.bashNonInteractive}" > "$path"

            for var in ${builtins.concatStringsSep " " envVars}; do
              echo "export $var" >> "$path"
            done

            echo -n 'exec ${icelock} ${icelockArgs} ' >> "$path"
            echo "-- "${config.app.package}$file" ${builtins.concatStringsSep " " config.app.flags} \"\$@\"" >> "$path"

            chmod +x "$path"
          done
        '';
      };

    };

}
