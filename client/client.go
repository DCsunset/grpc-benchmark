package main

import (
	"context"
	"flag"
	"fmt"
	"grpc-benchmark/api"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func main() {
	var address string
	var threadNum int
	flag.StringVar(&address, "address", "localhost:8900", "server address")
	flag.IntVar(&threadNum, "threads", 1, "num of threads")
	flag.Parse()

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Connection failed: ", err)
	}
	defer conn.Close()

	c := api.NewAPIClient(conn)
	ctx := context.Background()

	warmUp := 15
	expLength := 90
	coolDown := 15

	// errCount := make(map[string]int)
	errCount := 0
	var latencies = make([][]int64, threadNum)
	var counter = make([]int, threadNum)
	var wg sync.WaitGroup

	start := time.Now()

	for w := 0; w < threadNum; w++ {
		wg.Add(1)
		go func(id int) {
			for {
				if time.Since(start) >= time.Duration(expLength)*time.Second {
					break
				}
				s := time.Now()

				req := &api.Request{
					Data: "world",
				}

				resp, err := c.Call(ctx, req)
				if err != nil {
					fmt.Println("RPC failed: ", err)
					errCount++
					continue
				}
				fmt.Println("resp: ", resp.Data)

				e := time.Now()
				sinceStart := e.Sub(start)

				if sinceStart < time.Duration(warmUp)*time.Second || sinceStart > time.Duration(expLength-coolDown)*time.Second {
					continue
				}

				counter[id]++
				latency := e.Sub(s).Milliseconds()
				latencies[id] = append(latencies[id], latency)
			}

			wg.Done()
		}(w)
	}

	wg.Wait()
	var avgLatency float64 = 0
	latLen := 0
	for _, lat := range latencies {
		for _, l := range lat {
			avgLatency += float64(l)
			latLen++
		}
	}
	avgLatency /= float64(latLen)
	totalCounter := 0
	for _, cnt := range counter {
		totalCounter += cnt
	}

	f, err := os.OpenFile("result.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	throughputLog := fmt.Sprintf("throughput: %f\n", float64(totalCounter)/float64(expLength-warmUp-coolDown))
	latencyLog := fmt.Sprintf("latency: %f\n", avgLatency)
	errLog := fmt.Sprintf("errors: %d\n", errCount)

	f.WriteString(throughputLog)
	f.WriteString(latencyLog)
	f.WriteString(errLog)
}
