DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
source $DIR/preexec.sh

preexec () {
    echo $1 | curl -s -XGET --data @- http://localhost:8080/add/$$ > /dev/null
}

cleanup () {
   curl -s -XGET http://localhost:8080/delete/$$
}

trap 'cleanup' EXIT

_search() {
    res=$(echo $READLINE_LINE | curl -s -XGET --data @- http://localhost:8080/search/$$)
    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
}

_down() {
    res=$(echo $READLINE_LINE | curl -s -XGET --data @- http://localhost:8080/down/$$)
    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
}

_up() {
    res=$(echo $READLINE_LINE | curl -s -XGET --data @- http://localhost:8080/up/$$)
    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
}

bind -x '"\e[A": _up'
bind -x '"\e[B": _down'
bind -x '"\C-r": _search'
