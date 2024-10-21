package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/examples/require"
)

func main() {
	var remoteNode string
	flag.StringVar(&remoteNode, "node", "", "Addr of the target node host:port")
	var name string
	flag.StringVar(&name, "name", "?", "Session name")
	flag.Parse()

	ctx := context.Background()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if remoteNode != "" {
		fmt.Printf("To spawn a remote actor on %q\n", remoteNode)

		pid, err := goactor.Node().Spawn(ctx, remoteNode, &alice{
			SessionName: name,
		})
		require.NoError(err)

		fmt.Println("[!] Enter your message (CTRL+C to terminate)")
		inputCh := promptLoop()
		for {
			select {
			case msg := <-inputCh:
				err = goactor.Send(ctx, pid, msg)
				require.NoError(err, "could not send message to remote actor")
			case <-sigChan:
				return
			}
		}
	}

	goactor.Node().RegisterActorType(&alice{}, func() goactor.Actor {
		return &alice{
			SessionName: "none", // this will be replaced by the value the remote spawner provided
		}
	})
	_, port, _ := strings.Cut(goactor.Node().Addr(), ":")
	fmt.Printf("Registered node actor type 'Alice' on localhost port: %s\n", port)

	fmt.Println("[!] Press CTRL+C to terminate")
	<-sigChan
}

type alice struct {
	SessionName string
}

func (a *alice) Init(_ context.Context) error {
	fmt.Printf("[!] Alice %p spawned by %s\n", a, a.SessionName)
	return nil
}

func (a *alice) Receive(_ context.Context, msg any) error {
	fmt.Printf("[%s] %v\n", a.SessionName, msg)
	return nil
}

func promptLoop() <-chan string {
	ch := make(chan string)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")
		for s.Scan() {
			ch <- s.Text()
			fmt.Print("> ")
		}
	}()
	return ch
}
