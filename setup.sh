source $(dirname $BASH_SOURCE)/preexec.sh

preexec () {
    echo $1 | curl -s -XGET --data @- http://localhost:8080/add/$$ > /dev/null
}

cleanup () {
   curl -s -XGET http://localhost:8080/delete/$$
}

trap 'cleanup' EXIT
function clearLastLine() {
        tput cuu 1 && tput el
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
                JUUN_RES=$(echo $QUERY | curl -s -XGET --data @- http://localhost:8080/search/$$)
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
    res=$(echo $READLINE_LINE | curl -s -XGET --data @- http://localhost:8080/down/$$)

    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
    echo $READLINE_POINT
}

_up() {
    res=$(echo $READLINE_LINE | curl -s -XGET --data @- http://localhost:8080/up/$$)
    READLINE_LINE="$res"
    READLINE_POINT="${#READLINE_LINE}"
}

bind -x '"\e[A": _up'
bind -x '"\e[B": _down'
bind -x '"\C-p": _up'
bind -x '"\C-n": _down'

bind -x '"\C-r": "_search_start"'
