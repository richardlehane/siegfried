#!/usr/bin/env sh
set -ev # exit early on error
VERSION=$(echo "${TRAVIS_TAG}" | tr -d 'v')
BASE="${HOME}/deb"
SFDIR="${BASE}/siegfried_${VERSION}-1_amd64"
export SFDIR

# setup dirs
mkdir -p $SFDIR/DEBIAN
mkdir -p $SFDIR/usr/bin
mkdir -p $SFDIR/usr/share

# copy binaries and assets
cp $HOME/gopath/bin/sf $SFDIR/usr/bin/sf
cp $HOME/gopath/bin/roy $SFDIR/usr/bin/roy
cp -R $HOME/gopath/src/github.com/richardlehane/siegfried/cmd/roy/data/. $SFDIR/usr/share

# write control file
SIZE=$(du -s "${SFDIR}/usr" | cut -f1)
cat >$SFDIR/DEBIAN/control  << EOA
Package: siegfried
Version: $VERSION-1
Architecture: amd64
Maintainer: Richard Lehane <richard.lehane@gmail.com>
Installed-Size: $SIZE
Depends: libc6 (>= 2.2.5)
Section: misc
Priority: optional
Description: signature-based file identification tool
EOA

# make deb
cd $BASE
dpkg-deb --build $SFDIR

# write bintray json
DATE=`date +%Y-%m-%d`
cat >$BASE/bintray.json  << EOB
{
    "package": {
        "name": "siegfried",
        "repo": "debian",
        "subject": "siegfried",
        "desc": "see [CHANGELOG.md](https://github.com/richardlehane/siegfried/blob/master/CHANGELOG.md)",
        "website_url": "http://www.itforarchivists.com/siegfried",
        "issue_tracker_url": "https://github.com/richardlehane/siegfried/issues",
        "vcs_url": "hhttps://github.com/richardlehane/siegfried.git",
        "github_use_tag_release_notes": false,
        "licenses": ["Apache-2.0"],
        "public_download_numbers": false,
        "public_stats": false
    },

    "version": {
        "name": "${VERSION}",
        "desc": "Version ${VERSION}",
        "released": "${DATE}",
        "vcs_tag": "v${VERSION}",
        "gpgSign": false
    },

    "files":
        [
        {"includePattern": "${SFDIR}.deb", "uploadPattern": "\$1",
        "matrixParams": {
            "deb_distribution": "wheezy",
            "deb_component": "main",
            "deb_architecture": "amd64"}
        }
        ],
    "publish": false
}
EOB