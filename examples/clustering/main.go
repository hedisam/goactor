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
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/examples/lib/require"
	"github.com/hedisam/goactor/sysmsg"
)

func main() {
	var remoteNode string
	flag.StringVar(&remoteNode, "node", "", "Addr of the target node host:port")
	var name string
	flag.StringVar(&name, "name", "?", "Session name")
	var link bool
	flag.BoolVar(&link, "link", false, "whether spawn link or not")
	var trapExit bool
	flag.BoolVar(&trapExit, "trap-exit", false, "whether trap exit messages or not")
	flag.Parse()

	ctx := context.Background()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if remoteNode != "" {
		fmt.Printf("To spawn a remote actor on %q\n", remoteNode)

		pid, err := goactor.Node().Spawn(ctx, remoteNode, &alice{
			SessionName: name,
			TrapExit:    false,
		})
		require.NoError(err)
		linkerCtx, linkerCancel := context.WithCancel(ctx)
		if link {
			_, err = goactor.Spawn(linkerCtx, &linker{
				linkee:   pid,
				trapExit: trapExit,
			})
			require.NoError(err)
		}

		fmt.Println("[!] Enter your message (CTRL+C to terminate)")
		inputCh := promptLoop()
		for {
			select {
			case msg := <-inputCh:
				var n int
				_, err = fmt.Fscanf(strings.NewReader(msg), "bench send %d", &n)
				if err == nil {
					b, err := goactor.Node().Spawn(ctx, remoteNode, &bench{
						SessionName: name,
						Total:       n,
					})
					require.NoError(err)
					for range n {
						err = goactor.Send(ctx, b, msg)
						require.NoError(err)
					}
					continue
				}
				_, err = fmt.Fscanf(strings.NewReader(msg), "kill-linker")
				if err == nil {
					linkerCancel()
					continue
				}
				err = goactor.Send(ctx, pid, msg)
				require.NoError(err, "could not send message to remote actor")
			case <-sigChan:
				return
			}
		}
	}

	goactor.Node().RegisterActorType(&alice{}, func() goactor.Actor {
		return &alice{
			SessionName: "none", // this will be replaced by the value the remote spawner provides
		}
	})
	goactor.Node().RegisterActorType(&bench{}, func() goactor.Actor {
		return &bench{}
	})

	_, port, _ := strings.Cut(goactor.Node().Addr(), ":")
	fmt.Printf("Registered node actor type 'Alice' on localhost port: %s\n", port)

	fmt.Println("[!] Press CTRL+C to terminate")
	<-sigChan
}

type alice struct {
	SessionName string
	TrapExit    bool
}

func (a *alice) Init(_ context.Context) error {
	fmt.Printf("[!] Alice %p spawned by %s\n", a, a.SessionName)
	if a.TrapExit {
		fmt.Println("[!] Alice trapping exit messages")
		_ = goactor.SetTrapExit(true)
	}
	return nil
}

func (a *alice) Receive(_ context.Context, msg any) error {
	fmt.Printf("[%s] %v\n", a.SessionName, msg)
	return nil
}

type bench struct {
	SessionName string
	Total       int
	start       time.Time
	received    int
	self        *goactor.PID
}

func (b *bench) Init(_ context.Context) error {
	fmt.Printf("[!] Bench %p spawned by %s\n", b, b.SessionName)
	b.start = time.Now()
	b.self = goactor.Self()
	return nil
}

func (b *bench) Receive(_ context.Context, msg any) error {
	b.received++
	if b.received == b.Total {
		elapsed := time.Since(b.start)
		fmt.Printf("[!] Took %s to receive %d messages\n", elapsed, b.Total)
		fmt.Printf("[!] Message per second: %d\n", b.Total/int(elapsed.Seconds()))
		return sysmsg.ReasonNormal
	}
	return nil
}

type linker struct {
	linkee   *goactor.PID
	trapExit bool
}

func (l *linker) Init(_ context.Context) error {
	fmt.Println("[!] Linker spawned; linking to remote pid")
	if l.trapExit {
		fmt.Println("Trapping exit messages")
		_ = goactor.SetTrapExit(true)
	}
	return goactor.Link(l.linkee)
}

func (l *linker) Receive(ctx context.Context, msg any) error {
	fmt.Printf("[!] Linker recevied: %v\n", msg)
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
