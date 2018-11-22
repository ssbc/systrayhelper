with import <unstable> { };
systrayhelper.overrideDerivation (drv: { 
  name = "systrayhelper-fromsrc";
  src = ./.;
  buildFlagsArray = [ ''-ldflags=
    -X main.version=snap
    -X main.date="nix-byrev"
  '' ];
  goDeps = ./deps.nix;
  checkInputs = [ pkgs.xorg.xorgserver pkgs.i3 pkgs.i3status pkgs.xdotool pkgs.ffmpeg pkgs.x11vnc ];
  doCheck=true;
  checkPhase = ''
    export PATH=$PATH:$(pwd)/go/bin:${pkgs.xorg.xorgserver}/bin:${pkgs.i3status}/bin:${pkgs.i3}/bin:${pkgs.xdotool}/bin:${pkgs.ffmpeg}/bin:${pkgs.x11vnc}/bin

    ${pkgs.xorg.xorgserver}/bin/Xvfb ":23" -screen 0 800x600x16 &
    export DISPLAY=":23"

    # Wait for Xvfb
    MAX_ATTEMPTS=120 # About 60 seconds
    COUNT=0
    echo -n "Waiting for Xvfb to be ready..."
    while ! ${pkgs.xorg.xdpyinfo}/bin/xdpyinfo -display $DISPLAY; do
      echo -n "."
      sleep 0.5
      COUNT=$(( COUNT + 1 ))
      if [ "$COUNT" -ge "$MAX_ATTEMPTS" ]; then
        echo "  Gave up waiting for X server on $DISPLAY"
        exit 1
      fi
    done
    echo "  Done - Xvfb is ready!"

    export TRAY_XVFBRUNNING=t
    go test -timeout 5m -v github.com/ssbc/systrayhelper/test
  '';
})

