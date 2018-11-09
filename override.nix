with import <unstable> { };
systrayhelper.overrideDerivation (drv: { 
  name = "systrayhelper-fromsrc";
  src = ./.;
})

