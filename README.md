# JUUN - history forever

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

run the service

```
juun/juun.service
```

in some other terminal

```
source juun/setup.sh
```

this will hook up, down and ctrl+r to use juun for up/down history and search
it also hooks to preexec() (copied from https://github.com/rcaloras/bash-preexec) and every executed command goes into the service

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

* proper search
* run some basic ml to improve the search
* fix search "ui" to look more like readline's
* use protobuf for the history
* pick only one interface (now it listens on tcp 3333, unix socket on /tmp/juun.sock and http on 8080)


