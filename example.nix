{ pkgs, wrap, ... }:
{

  simple = wrap {
    app.package = pkgs.eza;

    fs.ro = [ "/" ];
  };

  limited-network = wrap {
    app.package = pkgs.curl;

    fs.ro = [ "/etc" ];

    net.connect = [ 443 ];

    seccomp.socketFamilies = [ "inet" ];

    mdwe = true;
  };

  gnome-calculator = wrap {
    app.package = pkgs.gnome-calculator;

    app.extraBinPaths = [ "/libexec/gnome-calculator-search-provider" ];

    fs.restrict = false;
  };

}
