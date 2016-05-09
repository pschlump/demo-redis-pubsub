package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/pubsub"
	"github.com/pschlump/radix.v2/redis" // Modified pool to have NewAuth for authorized connections

	"../qdemolib"
)

var Debug = flag.Bool("debug", false, "Debug flag")                      // 2
var Cfg = flag.String("cfg", "../global_cfg.json", "Configuration file") // 0
func init() {
	flag.BoolVar(Debug, "D", false, "Debug flag")                        // 2
	flag.StringVar(Cfg, "c", "../global_cfg.json", "Configuration file") // 0
}

func main() {

	flag.Parse()
	fns := flag.Args()

	if len(fns) != 0 {
		fmt.Printf("Usage: listen [ -c ../global_cfg.json ] [ -D ]\n")
		fmt.Printf("\t-c | --cfg    Configuraiton file name, default to ../global_cfg.json\n")
		fmt.Printf("\t-D | --debug  debug flag\n")
		os.Exit(1)
	}

	qdemolib.SetupRedisForTest(*Cfg)

	var wg sync.WaitGroup

	client, err := redis.Dial("tcp", qdemolib.ServerGlobal.RedisConnectHost+":"+qdemolib.ServerGlobal.RedisConnectPort)
	if err != nil {
		log.Fatal(err)
	}
	if qdemolib.ServerGlobal.RedisConnectAuth != "" {
		err = client.Cmd("AUTH", qdemolib.ServerGlobal.RedisConnectAuth).Err
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Fprintf(os.Stderr, "Success: Connected to redis-server with AUTH.\n")
		}
	} else {
		fmt.Fprintf(os.Stderr, "Success: Connected to redis-server.\n")
	}

	subClient := pubsub.NewSubClient(client)

	sr := subClient.Subscribe("listen")
	if sr.Err != nil {
		fmt.Fprintf(os.Stderr, "%sError: subscribe, %s.%s\n", MiscLib.ColorRed, sr.Err, MiscLib.ColorReset)
	}

	if sr.Type != pubsub.Subscribe {
		log.Fatalf("Did not receive a subscribe reply\n")
	}

	if sr.SubCount != 1 {
		log.Fatalf("Unexpected subscription count, Expected: 1, Found: %d\n", sr.SubCount)
	}

	// fmt.Printf("AT: %s\n", godebug.LF())

	subChan := make(chan *pubsub.SubResp)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			subChan <- subClient.Receive()
		}
	}()

	wg.Add(1)
	go func() {
		// fmt.Printf("AT: %s\n", godebug.LF())
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		defer wg.Done()
		for {
			// fmt.Printf("Before Receive AT: %s\n", godebug.LF())
			select {
			case sr = <-subChan:
				// sr={"Type":3,"Channel":"listen","Pattern":"","SubCount":0,"Message":"abcF","Err":null}
				fmt.Printf("***************** Got a message, sr=%+v\n", godebug.SVar(sr))
			case <-ticker.C:
				if *Debug {
					fmt.Printf("periodic ticker...\n")
				}
			}
			// fmt.Printf("After Receive AT: %s\n", godebug.LF())
		}
	}()

	// fmt.Printf("AT: %s\n", godebug.LF())
	wg.Wait()
}

/* vim: set noai ts=4 sw=4: */
