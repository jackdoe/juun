# JUUN - history forever

![logo](https://github.com/jackdoe/juun/raw/master/logo-small.png)


## learn damnit

![video](https://github.com/jackdoe/juun/raw/master/learn.gif)

in this example the search learns that by 'd' I mean `git diff`, not `dmesg`

## [Here be dragons](https://en.wikipedia.org/wiki/Here_be_dragons)

attempt to fix my bash history rage

* keeps my bash/zsh history across terminals
* local service that i can personalize and run some ML
* stores it in ~/.juun.json [0600]
* per terminal plus global history
+ has per terminal local state and when that is exhausted it uses the global history to go back

this annoys me so much in the current history(1) it is unbelivable

## install/run

### with curl

supported:
* macos amd64 [tested by me and some friends]
* linux amd64 [tested on travisci]
* freebsd amd64 [compiled, but not tested]

```
curl -L https://raw.githubusercontent.com/jackdoe/juun/master/download-and-install.sh | bash
```

### with homebrew:

```
brew tap jackdoe/tap
brew install juun
```
(then you need to follow the instructions)


### from source
requires bash4.+ or zsh, and golang

```
go get github.com/chzyer/readline
go get github.com/sevlyar/go-daemon

git clone https://github.com/jackdoe/juun
cd juun && make
```

```
make install # this will add 'source juun/dist/setup.sh' to .bash_profile and .zshrc
```

this will hook up, down and ctrl+r to use juun for up/down history and search
it also hooks to preexec() (bash: copied from https://github.com/rcaloras/bash-preexec) and every executed command goes into the service

setup.sh will always try to start `juun.service` which listens on $HOME/.juun.sock
logs are in $HOME/.juun.log and pid is $HOME/.juun.pid

## import

if you want to import your current history run:

```
$ HISTTIMEFORMAT= history | dist/juun.import
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

* [tfidf](https://en.wikipedia.org/wiki/Tf%E2%80%93idf) `(occurances of [m] in the line) * log(1-totalNumberDocuments/documentsHaving[m])`
* terminalScore `100 command was ran on this terminal session, 0 otherwise`
* countScore  `log(number of times this command line was executed)`
* timeScore `log10(seconds between now and the command)`
* score `tfidf + terminalScore + countScore + timeScore`

## learning

If you have [vowpal wabbit](https://github.com/VowpalWabbit/vowpal_wabbit) installed (`brew install vowpal-wabbit`), juun will use it to re-sort the last 5 items from the search
when you click (use) one of the recommended items it learns positive signal, if you use something else rather than the shown results, it will learn negative signal

Vowpal is started with quadratic interractions between `i` and `c` namespaces, the features are split into item and user features context, and the user context is `query` and the `time`.
For example: `git diff` is featurized as
```
|i_id id_2
|i_text git diff
|i_count count:4.454347
|i_time year_2018 day_25 month_12 hour_16
|i_score tfidf:1.870803 timeScore:-0.903090 countScore:4.454347 terminalScore_100
```
and the user is featurized as:

```
|c_user_time year_2018 day_25 month_12 hour_16
|c_query git
|c_cwd juun _Users_jack_work_juun
```

i_time is the last time this command was used, the idea is to learn patterns like: in the morning i prefer those commands, and in the evenening i prefer those
As you can see one of the features of the items is the search engine's score.


example log line in `~/.juun.log`

```
2018/12/25 16:54:26 sending 1 |i_id id_2  |i_text git diff  |i_count count:4.454347  |i_time year_2018 day_25 month_12 hour_16  |c_user_time year_2018 day_25 month_12 hour_16  |c_query git  |i_score tfidf:1.870803 timeScore:-0.903090 countScore:4.454347 terminalScore_100 |c_cwd juun _Users_jack_work_juun
2018/12/25 16:54:26 received 0.624512 0.584649 0.664374
```


## credit

logo: Icons made by <a href="https://www.freepik.com/" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" 			    title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" 			    title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a>

