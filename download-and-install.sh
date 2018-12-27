
os="darwin"
arch="amd64"

MACHINE_TYPE=$(uname -m)
if [ $MACHINE_TYPE != 'x86_64' ]; then
    echo "only 64 bit arch is supported at the moment"
    exit 1
fi

case $(uname -s) in
    Linux*)     os=linux;;
    Darwin*)    os=darwin;;
    FreeBSD*)   os=freebsd;;
    *)          echo "only mac/linux/freebsd are supported" && exit 1;;
esac

if [ -z "${JUUNURL}" ]; then
    JUUNURL=https://github.com/jackdoe/juun-bin-dist/raw/master/juun-latest-$os-$arch.tar.gz
fi

clean_exit () {
    [ ! -z "$LOCALINSTALLER" -a -f $LOCALINSTALLER ] && rm $LOCALINSTALLER
    exit $1
}

if [ -z "$TMPDIR" -o ! -d "$TMPDIR" ]; then
    if [ -d "/tmp" ]; then
        TMPDIR="/tmp"
    else
        TMPDIR="."
    fi
fi

cd $TMPDIR || clean_exit 1
LOCALINSTALLER=$(mktemp juun-install.XXXXXX)

echo

if type curl >/dev/null 2>&1; then
  JUUNDOWNLOAD="curl -f -sS -Lo $LOCALINSTALLER $JUUNURL"
elif type fetch >/dev/null 2>&1; then
  JUUNDOWNLOAD="fetch -o $LOCALINSTALLER $JUUNURL"
elif type wget >/dev/null 2>&1; then
  JUUNDOWNLOAD="wget -nv -O $LOCALINSTALLER $JUUNURL"
else
  echo "Need either wget, fetch or curl to use $0"
  clean_exit
fi

echo downloading $JUUNURL

$JUUNDOWNLOAD || clean_exit 1

echo

DEST=~/.juun.dist
mkdir $DEST

tar -C $DEST -xf ./$LOCALINSTALLER && $DEST/install.sh

rm ./$LOCALINSTALLER
