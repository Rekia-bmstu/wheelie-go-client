package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	wheeliegoclient "wheelie-client"
)

func main() {
	waitCh := make(chan os.Signal)
	signal.Notify(waitCh, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()
	client := wheeliegoclient.NewClient()
	err := client.Connect(ctx, "localhost", 9998, []byte("test-auth-data"))
	if err != nil {
		panic(err)
	}
	resp, err := client.Get(ctx, &wheeliegoclient.WheelieData{PktType: 77, Data: []byte("test-auth-data")})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	fireAndForgetPktStream := client.GetFireAndForgetChannel()
	go func() {
		for pkt := range fireAndForgetPktStream {
			fmt.Println(pkt)
		}
	}()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				if err := client.FireAndForget(ctx, wheeliegoclient.NewWheelieData(333, []byte("fire-and-forget"))); err != nil {
					panic(err)
				}
			}
		}
	}()

	serverPushCh := client.GetServerPushChannel()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				for pkt := range serverPushCh {
					fmt.Println(pkt)
				}
			}
		}
	}()
	<-waitCh
	if err := client.Close(ctx); err != nil {
		panic(err)
	}
}
