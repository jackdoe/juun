ROOT=$(dirname $BASH_SOURCE)
source $ROOT/preexec.sh

preexec () {
    work "add" "$1"
}

cleanup () {
    work delete $$
}

trap 'cleanup' EXIT

work() {
    $ROOT/juun.updown $1 $$ "$2"
}
_search_start() {
    $ROOT/juun.search $$ 2>/tmp/juun.search.$$
    res=$(cat /tmp/juun.search.$$ | tr -d "\n")
    rm /tmp/juun.search.$$


    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
}

_down() {
    res=$(work down $READLINE_LINE)
    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
}

_up() {
    res=$(work up $READLINE_LINE)
    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
}

bind -x '"\e[A": _up'
bind -x '"\e[B": _down'
bind -x '"\C-p": _up'
bind -x '"\C-n": _down'
bind -x '"\C-r": "_search_start"'

$ROOT/juun.service || "failed to start juun"
