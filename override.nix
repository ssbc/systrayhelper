with import ~/nixpkgs { };
systrayhelper.overrideDerivation (drv: { 
  name = "systrayhelper-fromsrc";
  src = ./.;
  buildFlagsArray = [ ''-ldflags=
    -X main.version=snap
  '' ];
  # use dep2nix to make this
  #goDeps = ./deps.nix;
})

