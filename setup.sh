ROOT=$(dirname $BASH_SOURCE)
source $ROOT/preexec.sh

preexec () {
    work "add" "$1"
}

cleanup () {
   curl --keepalive-time 60 -s -XGET http://localhost:8080/delete/$$
}

trap 'cleanup' EXIT

work() {
    # echo "$2" | curl --keepalive-time 60 -s -XGET --data @- http://localhost:8080/$1/$$
    # echo $1 $$ "$2" | nc -n 127.0.0.1 3333
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
