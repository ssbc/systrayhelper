with import <unstable> { };
systrayhelper.overrideDerivation (drv: { 
  name = "systrayhelper-fromsrc";
  src = ./.;
  buildFlagsArray = [ ''-ldflags=
    -X main.version=snap
    -X main.date="nix-byrev"
  '' ];
})

