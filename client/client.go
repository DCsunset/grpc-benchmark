package main

import (
	"context"
	"fmt"
	"grpc-benchmark/api"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
)

func main() {
	serverAddr := "localhost:8900"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Connection failed: ", err)
	}
	defer conn.Close()

	c := api.NewAPIClient(conn)
	ctx := context.Background()

	warmUp := 15
	expLength := 90
	coolDown := 15

	counter := 0
	// errCount := make(map[string]int)
	errCount := 0
	var latencies []int64

	start := time.Now()

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

		counter++
		latency := e.Sub(s).Milliseconds()
		latencies = append(latencies, latency)
	}

	var avgLatency float64 = 0
	for _, lat := range latencies {
		avgLatency += float64(lat)
	}
	avgLatency /= float64(len(latencies))

	args := os.Args[1:]
	name := ""
	if len(args) > 0 {
		name = fmt.Sprintf("-%s", args[0])
	}
	f, err := os.OpenFile(fmt.Sprintf("result%s.txt", name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	throughputLog := fmt.Sprintf("throughput: %f\n", float64(counter)/float64(expLength-warmUp-coolDown))
	latencyLog := fmt.Sprintf("latency: %f\n", avgLatency)
	errLog := fmt.Sprintf("errors: %d\n", errCount)

	f.WriteString(throughputLog)
	f.WriteString(latencyLog)
	f.WriteString(errLog)
}
