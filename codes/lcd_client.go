package main

import (
"flag"

"fmt"  //importing fmt package for printing



"math/rand" //importing for mathematical operations



"encoding/binary"

"github.com/scionproto/scion/go/lib/snet" //importing snet packages for the scion connections

"github.com/scionproto/scion/go/lib/sciond"
)

func geterror(e error){    //Error function

if e!=nil{

log.Println(e)

}

}
func main() {

var ( // necessary variable declarations

 clientadd string

 serveradd string

 packetsent string

 e error

 client *snet.Addr

 server *snet.Addr

 connectUDP *snet.Conn

)
flag.StringVar(&clientadd, "c", "", "address of client") // fetch address values from command line

flag.StringVar(&serveradd, "s", "", "address of server")// fetch server address from command line

flag.Parse()

client, e = snet.AddrFromString(clientadd)  // AddrFromString converts an address string of format isd-as,[ipaddr]:port


geterror(e)

server, e = snet.AddrFromString(serveradd)

geterror(e)

daddr := "/run/shm/dispatcher/default.sock"

flag.StringVar(&packetsent, "", message to be displayed)
flag.Parse()


_, e = connectUDP.Write(packetsent) //sending message to server
geterror(e)

}
