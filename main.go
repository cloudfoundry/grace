package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/grace/handlers"
	"github.com/cloudfoundry/grace/helpers"
	"github.com/cloudfoundry/grace/routes"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/rata"
)

func main() {
	var chatty bool
	var upFile string
	var exitAfter, startAfter time.Duration
	var exitAfterCode int
	var catchTerminate bool
	var attachToHostname bool

	flag.BoolVar(&chatty, "chatty", true, "make grace chatty")
	flag.StringVar(&upFile, "upFile", "", "a file to write to (lives under /tmp)")
	flag.DurationVar(&exitAfter, "exitAfter", 0, "if set, grace will exit after this duration")
	flag.IntVar(&exitAfterCode, "exitAfterCode", 0, "exit code to emit with exitAfter")
	flag.BoolVar(&catchTerminate, "catchTerminate", true, "make grace catch SIGTERM")
	flag.DurationVar(&startAfter, "startAfter", 0, "time to wait before starting")
	flag.BoolVar(&attachToHostname, "attachToHostname", false, "make grace attach to hostname:port instead of :port")

	flag.Parse()

	logger := lager.NewLogger("grace")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if startAfter > 0 {
		logger.Info("grace-is-sleeping-before-coming-up")
		time.Sleep(startAfter)
	}

	logger.Info("hello", lager.Data{"port": port})
	handler, err := rata.NewRouter(routes.Routes, handlers.New(logger))
	if err != nil {
		logger.Fatal("router.creation.failed", err)
	}

	if chatty {
		_, err := helpers.FetchIndex()

		go func() {
			t := time.NewTicker(time.Second)
			loop_index := 0
			for {
				<-t.C
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to fetch index: %s\n", err.Error())
				} else {
					// fmt.Printf("%s This is Grace on index: %d\n", time.Now().Format(time.RFC3339), index)
					for i := 0; i < 150; i++ {
						if loop_index >= 2 && loop_index < 5 {
							time.Sleep(500 * time.Millisecond)
							continue
						}
						if loop_index == 5 {
							fmt.Printf("index is 5")
						} else {
							fmt.Printf("%s This is Grace on index: %d-%d\n", time.Now().Format(time.RFC3339), i, loop_index)
						}
					}
					loop_index++
				}
			}
		}()
	}

	if upFile != "" {
		f, err := os.Create("/tmp/" + upFile)
		if err != nil {
			logger.Fatal("upfile.creation.failed", err)
		}
		_, err = f.WriteString("Grace is up")
		if err != nil {
			logger.Fatal("upfile.creation.failed", err)
		}
	}

	if exitAfter != 0 {
		go func() {
			time.Sleep(exitAfter)
			fmt.Println("timebomb... farewell")
			os.Exit(exitAfterCode)
		}()
	}

	if catchTerminate {
		go func() {
			fmt.Println(time.Now().Format(time.RFC3339), "Will Catch SIGTERM")
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGTERM)
			fmt.Println(time.Now().Format(time.RFC3339), "set up SIGTERM handler")
			<-c
			fmt.Println(time.Now().Format(time.RFC3339), "caught SIGTERM")
			for {
				fmt.Println(time.Now().Format(time.RFC3339), "still waiting to be killed...")
				time.Sleep(time.Second)
			}
		}()
	}

	addr := ":" + port
	secondaryPort := "9999"
	if sp := os.Getenv("SECONDARY_PORT"); sp != "" {
		secondaryPort = sp
	}
	secondary := ":" + secondaryPort
	if attachToHostname {
		hostname, err := os.Hostname()
		if err != nil {
			logger.Fatal("couldn't get hostname", err)
		}
		addr = hostname + ":" + port
		secondary = fmt.Sprintf("%s:%d", hostname, secondaryPort)
	}
	fmt.Printf("Grace is listening on %s\n", addr)

	server := ifrit.Invoke(grouper.NewOrdered(os.Interrupt,
		grouper.Members{
			{"primary", http_server.New(addr, handler)},
			{"secondary", http_server.New(secondary, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "grace side-channel")
			}))},
		},
	))

	go func() {
		//debug server
		debugPort := "6060"
		if dp := os.Getenv("DEBUG_PORT"); dp != "" {
			debugPort = dp
		}
		logger.Info("debug.server.starting", lager.Data{"port": debugPort})
		err := http.ListenAndServe("localhost:"+debugPort, nil)
		if err != nil {
			logger.Error("debug.server.failed", err)
		}
	}()

	err = <-server.Wait()
	if err != nil {
		logger.Error("farewell", err)
	}
	logger.Info("farewell")
}
