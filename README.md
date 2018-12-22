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
cd juun
go run main.go
```

in some other terminal

```
source juun/setup.sh
```

this will hook up, down and ctrl+r to use juun for up/down history and search
it also hooks to preexec() (copied from https://github.com/rcaloras/bash-preexec) and every executed command goes into the service

## todo

* proper search
* run some basic ml to improve the search
* fix search "ui" to look more like readline's



