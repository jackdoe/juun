if [[ -n "$BASH" ]]; then
    _realpath() {
        [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
    }
    ROOT=$(_realpath $(dirname $BASH_SOURCE))

    if [ ${BASH_VERSINFO[0]} -lt 4 ]; then
        echo "Sorry, you need at least bash-4.0 to use juun."
        exit 1
    fi

    source $ROOT/preexec.sh

    preexec () {
        work add "$1"
    }

    precmd () {
        work end end
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
        rc=$?
        res=$(cat /tmp/juun.search.$$)
        rm /tmp/juun.search.$$
        if [ $rc -eq 0 ]; then
            echo $res
            # FIXME: add it to the normal history?
            eval "$res"
            work "add" "$res"
        fi
        READLINE_LINE=""
        READLINE_POINT=""
    }

    _down() {
        res=$(work down down)
        READLINE_LINE="$res"
        READLINE_POINT="${#READLINE_LINE}"
    }

    _up() {
        res=$(work up up)
        READLINE_LINE="$res"
        READLINE_POINT="${#READLINE_LINE}"
    }

    bind -x '"\e[A": _up'
    bind -x '"\e[B": _down'
    bind -x '"\C-p": _up'
    bind -x '"\C-n": _down'
    bind -x '"\C-r": "_search_start"'

    $ROOT/juun.service || "failed to start juun"
elif [[ -n "$ZSH_VERSION" ]]; then
    ROOT=$(dirname $0:A)

    cleanup () {
        work delete $$
    }
    trap 'cleanup' EXIT

    work() {
        $ROOT/juun.updown $1 $$ "$2"
    }

    preexec () {
        work add "$1"
    }

    precmd () {
        work end end
    }

    _search_start() {
        zle -I
        </dev/tty $ROOT/juun.search $$ 2>/tmp/juun.search.$$
        rc=$?
        res=$(cat /tmp/juun.search.$$)
        rm /tmp/juun.search.$$

        if [ $rc -eq 0 ]; then
            BUFFER="$res"
            CURSOR=${#BUFFER}
            work "add" "$res"
            zle accept-line
        else
            BUFFER="$res"
            CURSOR=${#BUFFER}
        fi
    }

    _down() {
        BUFFER=$(work down down)
        CURSOR=${#BUFFER}
    }


    _up() {
        BUFFER=$(work up up)
        CURSOR=${#BUFFER}
    }

    zle -N _up
    zle -N _down
    zle -N _search_start

    bindkey "^[[A" _up
    bindkey "^[[B" _down
    bindkey "^p" _up
    bindkey "^n" _down
    bindkey "^R" _search_start

    $ROOT/juun.service || "failed to start juun"
else
    echo "Sorry, you need bash-4.+ or zsh to use juun."
    exit 1
fi
