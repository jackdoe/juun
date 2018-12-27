who_started_me() {
    # in case of curl .. | sh ..
    # the script is started by sh, but the PPID is actually whatever the user is using
    # so check if it is zsh, or bash and check the bash version
    comm=$(ps -o comm $PPID | tail -1)
    echo $comm | grep zsh > /dev/null 2>/dev/null
    rc=$?
    if [ $rc -eq 0 ]; then
        echo "zsh"
    else
        echo $comm | grep bash > /dev/null 2>/dev/null
        rc=$?
        if [ $rc -eq 0 ]; then
            # assume first bash in path is the one that started the script
            # maybe its better to check /etc/passwd for the shell?

            bash -version | grep "version 4" > /dev/null 2>/dev/null
            rc=$?
            if [ $rc -eq 0 ]; then
                echo "bash"
            else
                echo "unknown"
            fi
        else
            echo "unknown"
        fi
    fi
}

do_install() {
    fn=$1
    grep "$ROOT/setup.sh" $HOME/$fn > /dev/null 2>/dev/null
    rc=$?

    if [ $rc -eq 0 ]; then
        echo "already installed in $HOME/$fn"
    else
        echo "adding $ROOT/setup.sh to $HOME/$fn"
        echo source $ROOT/setup.sh >> $HOME/$fn
    fi

    which vw > /dev/null 2>/dev/null
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "you dont have VowpalWabbit installed, this means that juun will not be able to learn, on mac simply do brew install vowpal-wabbit, on linux you can apt-get/yum etc install it"
    else
        echo "found VowpalWabbit in $(which vw)"
    fi

    echo
    echo "run 'history | $ROOT/juun.import' in order to import your current history"
}



who=$(who_started_me)

echo "assuming $who as main shell"
echo

post_install() {
    echo
    echo "restarting juun.service from '$who'"
    echo
    
    $who -c "export JUUN_BIND_BASH=1 && source $ROOT/setup.sh && juun_restart"

    echo
    echo "done"
    echo
}

if [ "bash" = "$who" ]; then
    _realpath() {
        [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
    }
    ROOT=$(_realpath $(dirname $BASH_SOURCE))

    do_install ".bash_profile"
    post_install
elif [ "zsh" = "$who" ]; then
    ROOT=$(dirname $0:A)

    do_install ".zshrc"
    post_install
else
    echo "Sorry, only bash4+ and zsh are supported by juun"
    exit 1
fi
