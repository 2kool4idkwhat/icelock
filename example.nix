{ pkgs, makeIcelockWrapper, ... }:
{

  simple = makeIcelockWrapper {
    package = pkgs.eza;

    rx = [ "/nix/store" ];
    ro = [ "/" ];
  };

  limited-network = makeIcelockWrapper {
    package = pkgs.curl;

    rx = [ "/nix/store" ];
    ro = [ "/etc" ];

    connectTcp = [ 443 ];

    socketFamilies = [ "inet" ];
  };

  gnome-calculator = makeIcelockWrapper {
    package = pkgs.gnome-calculator;

    extraBinPaths = [ "/libexec/gnome-calculator-search-provider" ];

    socketFamilies = [ "unix" ];

    restrictFs = false;
  };

}
