{ pkgs, ... }:

let

  inherit (pkgs) lib;
  inherit (builtins) isBool isList;

  icelockPkg = pkgs.callPackage ./src { };
  icelock = lib.getExe icelockPkg;

  listOpt =
    flag: list: if list == [ ] then [ ] else "${flag} '${builtins.concatStringsSep "," list}'";

  portListOpt = flag: list: listOpt flag (map builtins.toString list);

  boolOpt = flag: value: if value then flag else [ ];
in

{
  makeIcelockWrapper =
    {
      package,
      extraBinPaths ? [ ],
      appFlags ? [ ],

      restrictFs ? true,
      ro ? [ ],
      rx ? [ "/nix/store" ],
      rw ? [ ],

      restrictNet ? true,
      bindTcp ? [ ],
      connectTcp ? [ ],

      scopeIpc ? true,

      seccompEnable ? true,
      seccompKill ? false,
      socketFamilies ? [ ],
      syscalls ? [ ],
    }:
    assert isBool restrictFs;
    assert isList ro;
    assert isList rx;
    assert isList rw;

    assert isBool restrictNet;
    assert isList bindTcp;
    assert isList connectTcp;

    assert isBool scopeIpc;

    assert isBool seccompEnable;
    assert isList socketFamilies;
    assert isList syscalls;
    let
      icelockArgs = builtins.concatStringsSep " " (
        lib.flatten [

          (boolOpt "--unrestricted-fs" (!restrictFs))
          (listOpt "--ro" ro)
          (listOpt "--rx" rx)
          (listOpt "--rw" rw)

          (boolOpt "--unrestricted-net" (!restrictNet))
          (portListOpt "--bind-tcp" bindTcp)
          (portListOpt "--connect-tcp" connectTcp)

          (boolOpt "--unscoped-ipc" (!scopeIpc))

          (boolOpt "--no-seccomp" (!seccompEnable))
          (boolOpt "--seccomp-kill" seccompKill)
          (listOpt "--af" socketFamilies)
          (listOpt "--syscalls" syscalls)
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
          echo "exec ${icelock} ${icelockArgs} -- "${package}/bin/$base" ${builtins.concatStringsSep " " appFlags} \$@" >> "$file"

          chmod +x "$file"
        done

        for file in ${builtins.concatStringsSep " " extraBinPaths}; do
          path="$out/$file"

          echo "wrapping $path"
          rm "$path"

          echo "#!${lib.getExe pkgs.bashNonInteractive}" > "$path"
          echo "exec ${icelock} ${icelockArgs} -- "${package}$file" ${builtins.concatStringsSep " " appFlags} \$@" >> "$path"

          chmod +x "$path"
        done
      '';
    };
}
