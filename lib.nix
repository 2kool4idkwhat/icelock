{ pkgs, ... }:

let

  inherit (pkgs) lib;
  inherit (builtins) isBool isList;

  icelockPkg = pkgs.callPackage ./package.nix { };
  icelock = lib.getExe icelockPkg;

  listOpt =
    flag: list: if list == [ ] then [ ] else ''${flag}="${builtins.concatStringsSep ''","'' list}"'';

  portListOpt = flag: list: listOpt flag (map builtins.toString list);

  boolOpt = flag: value: if value then flag else [ ];
in

{
  makeIcelockWrapper =
    {
      package,
      extraBinPaths ? [ ],
      appFlags ? [ ],
      env ? { },

      restrictFs ? true,
      ro ? [ ],
      rx ? [ "/nix/store" ],
      rw ? [ ],
      unix ? [ ],

      restrictNet ? true,
      bind ? [ ],
      bindAll ? false,
      connect ? [ ],
      connectAll ? false,

      signals ? false,
      abstractUnixSockets ? false,

      seccompEnable ? true,
      socketFamilies ? [ ],
      syscalls ? [ ],

      keepCaps ? false,

      userNamespaces ? false,
      ioUring ? false,

      mdwe ? false,
    }:
    assert isBool restrictFs;
    assert isList ro;
    assert isList rx;
    assert isList rw;
    assert isList unix;

    assert isBool restrictNet;
    assert isList bind;
    assert isList connect;

    assert isBool seccompEnable;
    assert isList socketFamilies;
    assert isList syscalls;

    assert isBool userNamespaces;
    assert isBool ioUring;

    assert isBool keepCaps;

    assert isBool mdwe;
    let
      envVars = lib.mapAttrsToList (name: value: ''${name}='"${value}"' '') env;

      icelockArgs = builtins.concatStringsSep " " (
        lib.flatten [

          (boolOpt "--unrestricted-fs" (!restrictFs))
          (listOpt "--ro" ro)
          (listOpt "--rx" rx)
          (listOpt "--rw" rw)
          (listOpt "--unix" unix)

          (boolOpt "--unrestricted-net" (!restrictNet))
          (portListOpt "--bind" bind)
          (boolOpt "--bind-all" bindAll)
          (portListOpt "--connect" connect)
          (boolOpt "--connect-all" connectAll)

          (boolOpt "--signals" signals)
          (boolOpt "--abstract-unix" abstractUnixSockets)

          (boolOpt "--no-seccomp" (!seccompEnable))
          (listOpt "--af" socketFamilies)
          (listOpt "--syscalls" syscalls)

          (boolOpt "--userns" userNamespaces)
          (boolOpt "--io-uring" ioUring)

          (boolOpt "--keep-caps" keepCaps)

          (boolOpt "--mdwe" mdwe)
        ]
      );
    in

    pkgs.symlinkJoin {
      name = package.name;
      paths = [ package ] ++ (if (builtins.hasAttr "man" package) then [ package.man ] else [ ]);
      passthru.unwrapped = package;
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
          echo "-- "${package}/bin/$base" ${builtins.concatStringsSep " " appFlags} \"\$@\"" >> "$file"

          chmod +x "$file"
        done

        for file in ${builtins.concatStringsSep " " extraBinPaths}; do
          path="$out/$file"

          echo "wrapping $path"
          rm "$path"

          echo "#!${lib.getExe pkgs.bashNonInteractive}" > "$path"

          for var in ${builtins.concatStringsSep " " envVars}; do
            echo "export $var" >> "$path"
          done

          echo -n 'exec ${icelock} ${icelockArgs} ' >> "$path"
          echo "-- "${package}$file" ${builtins.concatStringsSep " " appFlags} \"\$@\"" >> "$path"

          chmod +x "$path"
        done
      '';
    };
}
