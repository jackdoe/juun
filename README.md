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

### with homebrew:

```
brew tap jackdoe/tap
brew install juun
```
(then you need to follow the instructions)


### from source
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

## import

if you want to import your current history run:

```
$ history | dist/juun.import
```

this will add each of your history lines to juun

## scoring

running search for `m` from one terminal gives the following score
(edge gram `e_m`)

```

2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.162952 terminalScore:0.000000 countScore:4.543295, age: 14552s - make
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.904991 terminalScore:0.000000 countScore:1.386294, age: 80350s - make -n
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.820044 terminalScore:0.000000 countScore:0.693147, age: 66075s - make clean
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.757343 terminalScore:0.000000 countScore:1.945910, age: 57192s - make install
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.754386 terminalScore:0.000000 countScore:0.693147, age: 56804s - mkdir brew
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.114077 terminalScore:0.000000 countScore:0.693147, age: 13003s - git commit -a -m 'make it feel more natural; fix issues with newline'

```

* tfidf `(occurances of [m] in the line) * log(1-totalNumberDocuments/documentsHaving[m])`
* terminalScore `100 command was ran on this terminal session, 0 otherwise`
* countScore  `log(number of times this command line was executed)`
* timeScore `log10(seconds between now and the command)`
* score `tfidf + terminalScore + countScore + timeScore`


## todo

* run some basic ml to improve the search
* fix search "ui" to look more like readline's
* use protobuf for the history
* limit the amount of data in the indexes (archive)

## cerdit

logo: <div>Icons made by <a href="https://www.freepik.com/" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" 			    title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" 			    title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a></div>

