{ stdenv, pkgconfig, gtk3, libappindicator-gtk3, buildGoPackage, fetchgit, fetchhg, fetchbzr, fetchsvn }:

buildGoPackage rec {
  name = "systray-jsonhelper-unstable-${version}";
  version = "2018-07-08";
  rev = "99cc53acaf2e9287ff8a5f0abab3d1fc60b4bc55";

  goPackagePath = "github.com/ssbc/systray-jsonhelper";

  src = fetchgit {
    inherit rev;
    url = "git@github.com:ssbc/systray-jsonhelper.git";
    sha256 = "0iwq7808zq1jlgficmk0xmi38my1a46r5r9wd5nzgvdc0vprw0sb";
  };

  goDeps = ./deps.nix;

  nativeBuildInputs = [ pkgconfig gtk3 libappindicator-gtk3 ];
  # https://blog.kiloreux.me/2018/05/24/learning-nix-by-example-building-ffmpeg-4-dot-0/
  # o/ says not `buildInputs` are used by the code at run-time, so no


  meta = with stdenv.lib; {
    description = "A portable version of go systray, using stdin/stdout to communicate with other language";
    homepage    = "https://github.com/ssbc/systray-jsonhelper";
    maintainers = with maintainers; [ cryptix ];
    license     = licenses.mit;
    # It depends on the inputs, i guess? not sure about solaris, for instance. go supports it though
    # I hope nix can figure this out?! ¯\\_(ツ)_/¯
    platforms   = platforms.all;
  };
}
