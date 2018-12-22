source $(dirname $BASH_SOURCE)/preexec.sh

preexec () {
    work "add" "$1"
}

cleanup () {
   curl --keepalive-time 60 -s -XGET http://localhost:8080/delete/$$
}

trap 'cleanup' EXIT
function clearLastLine() {
        tput cuu 1 && tput el
}
work() {
     echo "$2" | curl --keepalive-time 60 -s -XGET --data @- http://localhost:8080/$1/$$
    # echo $1 $$ "$2" | nc -n 127.0.0.1 3333
}
_search_start() {
    QUERY=""
    JUUN_RES=""
    QUERY=$READLINE_LINE
    SEARCHING=1
    JP="juun> "

    echo -n "$JP"
    POINT=$READLINE_POINT
    while read -e -s -p '' -n1 c; do
        case $c in
            "")
                break
                ;;
            *)
                clearLastLine
                QUERY="$QUERY$c"
                JUUN_RES=$(work search $QUERY)
                if [ "$JUUN_RES" = "" ]; then
                    echo -en "$JP$QUERY";
                else
                    echo -en "$JP$JUUN_RES";
                fi
                ;;
        esac
    done

    clearLastLine

    out=""
    if [ "$JUUN_RES" = "" ]; then
        out="$QUERY";
    else
        out="$JUUN_RES";
    fi
    eval $out
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
