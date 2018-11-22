with import <unstable> { };
systrayhelper.overrideDerivation (drv: { 
  name = "systrayhelper-fromsrc";
  src = ./.;
  buildFlagsArray = [ ''-ldflags=
    -X main.version=snap
    -X main.date="nix-byrev"
  '' ];
  goDeps = ./deps.nix;
  checkInputs = [ pkgs.xorg.xorgserver pkgs.i3 pkgs.i3status pkgs.xdotool pkgs.ffmpeg ];
  doCheck=true;
  checkPhase = ''
    export PATH=$PATH:${pkgs.xorg.xorgserver}/bin:${pkgs.i3status}/bin:${pkgs.i3}/bin:${pkgs.xdotool}/bin:${pkgs.ffmpeg}/bin:$(pwd)/go/bin
    ${pkgs.xorg.xorgserver}/bin/Xvfb ":23" &
    export DISPLAY=":23"
    export TRAY_RECORD=/tmp/recording.mp4
    go test -timeout 1m -v github.com/ssbc/systrayhelper/test
  '';
})

