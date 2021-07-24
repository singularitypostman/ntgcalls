package main
import "C"
import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var joinCallResult map[int]chan string
var tgCallResponse map[int]chan string

func main() {}

//export initClient
func initClient(){
	joinCallResult = make(map[int]chan string)
	tgCallResponse = make(map[int]chan string)
	//fmt.Println(os.Environ())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Fatal("Closed")
	}()
}