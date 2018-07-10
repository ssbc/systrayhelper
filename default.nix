{ stdenv, pkgconfig, gtk3, libappindicator-gtk3, buildGoPackage, fetchgit, fetchhg, fetchbzr, fetchsvn }:

buildGoPackage rec {
  name = "systrayhelper-unstable-${version}";
  version = "2018-07-10";
  rev = "552edee02edda55d077425e17dc6f3036331a740";

  goPackagePath = "github.com/ssbc/systrayhelper";

  src = fetchgit {
    inherit rev;
    url = "https://github.com/ssbc/systrayhelper.git";
    sha256 = "0609f7lbdgvfcsx8a7zg0idxhjja43jni5s7718gs5n23m3znrsz";
  };

  goDeps = ./deps.nix;

  nativeBuildInputs = [ pkgconfig gtk3 libappindicator-gtk3 ];
  # https://blog.kiloreux.me/2018/05/24/learning-nix-by-example-building-ffmpeg-4-dot-0/
  # o/ says not `buildInputs` are used by the code at run-time, so no

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
