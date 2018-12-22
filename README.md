# JUUN - history forever

![logo](https://github.com/jackdoe/juun/raw/master/logo.png)

## work in progress

attempt to fix my bash history rage

* keeps my bash history across terminals
* local service that i can personalize and run some ML
* stores it in ~/.juun.json [0600]
* per terminal plus global history
+ has per terminal local state and when that is exhausted it uses the global history to go back

this annoys me so much in the current history(1) it is unbelivable



## install/run

requires bash4.+ and golang

```
git clone https://github.com/jackdoe/juun
cd juun && make
```

```
make install # this will add 'source juun/dist/setup.sh' to .bash_profile
```

this will hook up, down and ctrl+r to use juun for up/down history and search
it also hooks to preexec() (copied from https://github.com/rcaloras/bash-preexec) and every executed command goes into the service

setup.sh will always try to start `juun.service` which listens on $HOME/.juun.sock
logs are in $HOME/.juun.log and pid is $HOME/.juun.pid

## scoring

running search for `m` from one terminal gives the following score

```
tfidf: 0.693147 timeScore: -11.421828 terminalScore:100.000000 make
tfidf: 0.693147 timeScore: -11.416780 terminalScore:100.000000 mono
tfidf: 0.693147 timeScore: -11.065346 terminalScore:0.000000 mongo
```
the word mongo is typed on another terminal, and so it is downscored
the time score is `-log10(now - command.timestamp)`



## todo

* run some basic ml to improve the search
* fix search "ui" to look more like readline's
* use protobuf for the history
* limit the amount of data in the indexes (archive)

## cerdit

logo: <div>Icons made by <a href="https://www.freepik.com/" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" 			    title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" 			    title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a></div>

