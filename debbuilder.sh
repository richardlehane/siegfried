#!/usr/bin/env sh
set -ev # exit early on error

# setup dirs
mkdir -p $SF_PATH/DEBIAN
mkdir -p $SF_PATH/usr/bin
mkdir -p $SF_PATH/usr/share/siegfried

ls $SF_PATH

# copy binaries and assets
cp $BIN_PATH/sf $SF_PATH/usr/bin/
cp $BIN_PATH/roy $SF_PATH/usr/bin/
cp -R cmd/roy/data/. $SF_PATH/usr/share/siegfried

# write control file
SIZE=$(du -s "${SF_PATH}/usr" | cut -f1)
cat >$SF_PATH/DEBIAN/control  << EOA
Package: siegfried
Version: $VERSION-1
Architecture: amd64
Maintainer: Richard Lehane <richard.lehane@itforarchivists.com>
Installed-Size: $SIZE
Depends: libc6 (>= 2.2.5)
Section: misc
Priority: optional
Description: signature-based file identification tool
EOA

# make deb; explicit 'xz' is for compatibility with Debian "bullseye";
# see:
#
#    https://github.com/richardlehane/siegfried/issues/222
#
dpkg-deb -Zxz --build $SF_PATH
