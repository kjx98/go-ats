# goATS
[![Build Status](https://api.travis-ci.org/kjx98/go-ats.svg?branch=master)](
https://api.travis-ci.org/kjx98/go-ats)

open Algorithm Trading Platform using golang

goATS inspired by abstract Broker/Feed from pyalgotrade, Order Type&Execution from InteractiveBrokers.

goATS is a abstract Broker/Feed service Library for go/lua/Python/Java Algorithm Tradin System. goATS works great with the pyalgotrade open source backtesting library.

The main function of goATS is to providing unified broker&feed service for backtesting & trading(real or paper), including:

-  simulation tick from Bar
-  backtesting with tick data
-  realtime quotes feed
-  history data feed, directly from feed or local cached
-  paper or real trading with broker

### Documentation

Visit the docs on [gopkgdoc] go-ats (http://godoc.org/github.com/kjx98/go-ats)
