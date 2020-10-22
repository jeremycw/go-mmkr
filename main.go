package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type joinRequest struct {
	Score int `json:"score"`
}

type joinResponse struct {
	Uuid string `json:"id"`
}

type matchConfig struct {
	minMatchSize   int
	maxMatchSize   int
	matchTimeoutMs int
}

func matchServer(channel chan matchCmd, conf matchConfig) {
	mmkr := newMatchMaker(conf.minMatchSize, conf.maxMatchSize, conf.matchTimeoutMs)
	for cmd := range channel {
		cmd.exec(mmkr)
	}
}

func postJoin(channel chan matchCmd, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var join joinRequest
	decoder.Decode(&join)
	id := uuid.New()
	cmd := new(joinCmd)
	*cmd = joinCmd{id: id, score: join.Score}
	channel <- cmd
	resp := joinResponse{Uuid: id.String()}
	body, err := json.Marshal(&resp)
	if err != nil {
		return
	}
	w.Write(body)
}

func getMatch(channel chan matchCmd, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("session_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request - Malformed Session Id"))
		return
	}
	resChan := make(chan uuid.UUID)
	cmd := new(watchCmd)
	*cmd = watchCmd{id: id, channel: resChan}
	channel <- cmd
	matchId := <-resChan
	if matchId == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
		return
	}
	body, err := json.Marshal(joinResponse{Uuid: matchId.String()})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error - Failed to Serialize JSON"))
		return
	}
	w.Write(body)
}

func tickMatchMaker(channel chan matchCmd, tickMs int) {
	for {
		duration := time.Duration(tickMs) * time.Millisecond
		time.Sleep(duration)
		cmd := new(tickCmd)
		*cmd = tickCmd{delta: tickMs}
		channel <- cmd
	}
}

func main() {
	channel := make(chan matchCmd)
	minSize := flag.Int("min-size", 2, "Minimum amount of users required for a match")
	maxSize := flag.Int("max-size", 32, "Maximum amount of users for a match")
	timeout := flag.Int("timeout", 30000, "Amount of time in ms to wait for match")
	period := flag.Int("process-period", 1000, "Amount of time in ms to wait between computing match-ups")
	port := flag.Int("port", 8080, "Port to bind to")
	flag.Parse()
	conf := matchConfig{
		minMatchSize:   *minSize,
		maxMatchSize:   *maxSize,
		matchTimeoutMs: *timeout,
	}
	go matchServer(channel, conf)
	go tickMatchMaker(channel, *period)
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postJoin(channel, w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}
	})
	http.HandleFunc("/match", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getMatch(channel, w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
