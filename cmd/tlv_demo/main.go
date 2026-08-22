package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
	wheeliegoclient "wheelie-client"
)

func main() {
	ctx := context.Background()

	tlv := wheeliegoclient.NewTlv("localhost", 9998)

	err := tlv.Dial(ctx)
	if err != nil {
		panic(err)
	}

	waitCh := make(chan os.Signal, 1)
	signal.Notify(waitCh, syscall.SIGINT, syscall.SIGTERM)
	t := time.NewTicker(time.Second)

	receivech, doneErr, err := tlv.StartReceive(ctx)
	go func() {
		for val := range receivech {
			fmt.Printf("tag: %d, string data: %s\n", val.Tag, string(val.Data))
		}
	}()
	go func() {
		err := <-doneErr
		fmt.Println("doneErr:", err)
	}()

	for {
		select {
		case <-waitCh:
			return
		case <-t.C:
			tag := rand.Int32()
			r := rand.Int32()
			err := tlv.Send(ctx, tag, []byte(fmt.Sprintf("hello world %d", r)))
			if err != nil {
				fmt.Println("send error:", err)
			}
		}
	}
}
