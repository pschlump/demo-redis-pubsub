package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/pubsub"
	"github.com/pschlump/radix.v2/redis" // Modified pool to have NewAuth for authorized connections
)

func main() {

	SetupRedisForTest("../global_cfg.json")

	/*
		// var client *redis.Client
		client, err = ServerGlobal.RedisPool.Get()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError: Unable to get connection to redis-server.%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		}
		defer ServerGlobal.RedisPool.Put(client)
	*/
	var wg sync.WaitGroup

	// client, err := redis.DialTimeout("tcp", ServerGlobal.RedisConnectHost+":"+ServerGlobal.RedisConnectPort, time.Duration(10)*time.Second)
	client, err := redis.Dial("tcp", ServerGlobal.RedisConnectHost+":"+ServerGlobal.RedisConnectPort)
	if err != nil {
		log.Fatal(err)
	}
	if ServerGlobal.RedisConnectAuth != "" {
		err = client.Cmd("AUTH", ServerGlobal.RedisConnectAuth).Err
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

	fmt.Printf("AT: %s\n", godebug.LF())

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
		fmt.Printf("AT: %s\n", godebug.LF())
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		defer wg.Done()
		for {
			fmt.Printf("Before Receive AT: %s\n", godebug.LF())
			select {
			case sr = <-subChan:
				// sr={"Type":3,"Channel":"listen","Pattern":"","SubCount":0,"Message":"abcF","Err":null}
				fmt.Printf("***************** Got a message, sr=%+v\n", godebug.SVar(sr))
			case <-ticker.C:
				fmt.Printf("periodic ticker...\n")
			}
			fmt.Printf("After Receive AT: %s\n", godebug.LF())
		}
	}()

	fmt.Printf("AT: %s\n", godebug.LF())
	wg.Wait()
}

/*
		// quit := make(chan struct{})
			//select {
			//case sr = <-subChan:
			//	fmt.Printf("***************** Got a message, subChan=%+v\n", sr)
			//case <-ticker.C:
			//	fmt.Printf("Took too long to Receive message\n")
			//case <-quit:
			//	fmt.Printf("That's all folks...\n")
			//	return
			//}

	r := pub.Cmd("PUBLISH", channel, message)
	if r.Err != nil {
		t.Fatal(r.Err)
	}


	select {
	case sr = <-subChan:
	case <-time.After(time.Duration(10) * time.Second):
		t.Fatal("Took too long to Receive message")
	}

	if sr.Err != nil {
		t.Fatal(sr.Err)
	}

	if sr.Type != Message {
		t.Fatal("Did not receive a message reply")
	}

	if sr.Message != message {
		t.Fatal(fmt.Sprintf("Did not recieve expected message '%s', instead got: '%s'", message, sr.Message))
	}

	sr = sub.Unsubscribe(channel)
	if sr.Err != nil {
		t.Fatal(sr.Err)
	}

	if sr.Type != Unsubscribe {
		t.Fatal("Did not receive a unsubscribe reply")
	}

	if sr.SubCount != 0 {
		t.Fatal(fmt.Sprintf("Unexpected subscription count, Expected: 0, Found: %d", sr.SubCount))
	}
}

*/

/* vim: set noai ts=4 sw=4: */
