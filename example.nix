{ pkgs, makeIcelockWrapper, ... }:
{

  simple = makeIcelockWrapper {
    package = pkgs.eza;

    ro = [ "/" ];
  };

  limited-network = makeIcelockWrapper {
    package = pkgs.curl;

    ro = [ "/etc" ];

    connect = [ 443 ];

    socketFamilies = [ "inet" ];

    mdwe = true;
  };

  gnome-calculator = makeIcelockWrapper {
    package = pkgs.gnome-calculator;

    extraBinPaths = [ "/libexec/gnome-calculator-search-provider" ];

    restrictFs = false;
  };

}
