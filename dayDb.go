package ats

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/kjx98/golib/julian"
)

// Symbol
//	symbol table for assets
//	StartDate, EndDate,	AutoCloseDate julianDay
type Symbol struct {
	Ticker        string
	Name          string
	StartDate     julian.JulianDay
	EndDate       julian.JulianDay
	AutoCloseDate julian.JulianDay
	Exchange      string
	bNew          bool
}

var symbolsMap = map[string]*Symbol{}

func getTickerExchange(ticker string) string {
	switch ticker[:2] {
	case "SH", "sh":
		return "SSE"
	case "SZ", "sz":
		return "SZSE"
	}
	return "SSE"
}

func FindTicker(ticker string) *Symbol {
	if sym, ok := symbolsMap[ticker]; ok {
		return sym
	}
	exch := getTickerExchange(ticker)
	var sym = Symbol{Ticker: ticker, Exchange: exch, StartDate: 0,
		EndDate: 0, bNew: true}
	symbolsMap[ticker] = &sym
	return &sym
}

func IsEquity(code string) int {
	switch string(code[:4]) {
	case "sh60":
		return 2 // SHSE A
	case "sh00":
		return 1 // SHSE Indexes
	case "sz00":
		return 2 // SZSE A
	case "sz30":
		return 2 // SZSE GEM growth enterprise market
	case "sh51":
		return 3 // SHSE ETF
		//case "sz20": return 3 // SZSE B
		//case "sh90": return 3 // SHSE B
	}
	switch string(code[:5]) {
	case "sz159":
		return 3 // SZSE ETF
	case "sz399":
		return 1 // SZSE Indexes
	}
	return 0
}

func GetDB() *sql.DB {
	return myDB
}

var myDB *sql.DB

func OpenDB() (*sql.DB, error) {
	if myDB != nil {
		return myDB, nil
	}
	if db, err := sql.Open("mysql", "/tadb?charset=gbk"); err != nil {
		return nil, err
	} else {
		myDB = db
	}
	// init symbolsMap
	rows, err := myDB.Query("select * from symbols")
	if err == nil {
		for rows.Next() {
			var tick, name, start, end, autoc, exch string
			err := rows.Scan(&tick, &name, &start, &end, &autoc, &exch)
			if err != nil {
				panic(err)
			}
			var sym = Symbol{Ticker: tick, Name: name,
				StartDate: julian.FromString(start),
				EndDate:   julian.FromString(end), Exchange: exch}
			symbolsMap[tick] = &sym
		}
	} else {
		myDB = nil
		return nil, err
	}
	return myDB, nil
}

var cacheDayHits int
var cacheDayMiss int

func DayDbCacheStatus() string {
	return fmt.Sprintf("DayTACache Status: Hits %d, Miss: %d", cacheDayHits, cacheDayMiss)
}

func GetChart(sym string, startD, endD julian.JulianDay) (res []DayTA) {
	if cc, ok := cacheDayTA[sym]; ok {
		if startD == cc.startD && endD == cc.endD {
			res = cc.res
			cacheDayHits++
			return
		} else {
			cacheDayMiss++
		}
	}
	si, err := GetSymbolInfo(sym)
	if err != nil {
		fmt.Println("GetSymbolInfo err", err)
		return
	}
	dMulti := digitMulti(si.PriceDigits)
	rows, err := myDB.Query("select * from day_ta where code='" +
		sym + "' order by ta_time")
	if err != nil {
		fmt.Println("myDB query err", err)
	} else {
		var Day string
		var Open, High, Low, Close float64
		var Volume int64
		var Turnover float32
		var symbol string
		for rows.Next() {
			err := rows.Scan(&symbol, &Day, &Open, &High, &Low, &Close,
				&Volume, &Turnover)
			if err != nil {
				fmt.Printf("mysql Scan error: %v", err)
				return
			}
			if symbol != sym {
				fmt.Printf("mysql query err. sym: %s != select code: %s\n",
					sym, symbol)
				return
			}
			var dayDD = DayTA{Date: julian.FromString(Day), Open: int32(Open * dMulti),
				High: int32(High * dMulti), Low: int32(Low * dMulti),
				Close: int32(Close * dMulti), Volume: Volume, Turnover: Turnover}
			if dayDD.Date < startD {
				continue
			}
			if endD != 0 && dayDD.Date > endD {
				continue
			}
			res = append(res, dayDD)
		}
	}
	var cc = cacheDayTAType{startD: startD, endD: endD}
	cc.res = res
	cacheDayTA[sym] = cc
	return
}

func init() {
	if _, err := OpenDB(); err != nil {
		log.Warning("init dayDb, OpenDB()", err)
	}
}
