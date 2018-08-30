{ stdenv, pkgconfig, gtk3, libappindicator-gtk3, buildGoPackage, fetchFromGitHub }:

buildGoPackage rec {
  name = "systrayhelper-${version}";
  version = "0.0.1";
  rev = "94ce96498bc55d5be01c8d5e51742eae54e41e79";

  goPackagePath = "github.com/ssbc/systrayhelper";

  src = fetchFromGitHub {
    inherit rev;
    owner = "ssbc";
    repo = "systrayhelper";
    sha256 = "1kv9hq194b0zlzfd3k7274kgarbs9lijirqn0f441cmq3ssk8wgh";
  };

  goDeps = ./deps.nix;

  # -X main.date=${date +%F}?
  buildFlagsArray = [ ''-ldflags=
    -X main.version=v${version}
    -X main.commit=${rev}
    -s
    -w
  '' ];

  nativeBuildInputs = [ pkgconfig gtk3 ];
  buildInputs = [ libappindicator-gtk3 ];

  meta = with stdenv.lib; {
    description = "A portable version of go systray, using stdin/stdout to communicate with other language";
    homepage    = "https://github.com/ssbc/systrayhelper";
    maintainers = with maintainers; [ cryptix ];
    license     = licenses.mit;
    # It depends on the inputs, i guess? not sure about solaris, for instance. go supports it though
    # I hope nix can figure this out?! ¯\\_(ツ)_/¯
    platforms   = platforms.all;
  };
}
