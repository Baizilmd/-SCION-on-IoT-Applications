//lcd tutorail http://www.microcontrollerboard.com/lcd.html
//https://electronicsforu.com/resources/learn-electronics/16x2-lcd-pinout-diagram
//https://electronicsforu.com/resources/learn-electronics/16x2-lcd-pinout-diagram
package main

import (
	"github.com/stianeikeland/go-rpio"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"log"
	"strings"
	"sync"
	"time"

	"flag"

	"fmt"  //importing fmt package for printing



	"math/rand" //importing for mathematical operations



  "encoding/binary"

	"github.com/scionproto/scion/go/lib/snet" //importing snet packages for the scion connections

	"github.com/scionproto/scion/go/lib/sciond"
)

//Display (output) interface
type Display interface {
	Display(string)
	Close()
}

const (
	//Timing constants
	toggle = 1 * time.Microsecond // alloting pulse of 1 millisecond for toggle operation
	delay = 50 * time.Microsecond // creating delay of 50 millisecond

	//broadcom pin nos or GPIO pins
	lcdRS = 6 //RS is register select pin and is in instruction/commands mode(RS=0) and data mode(RS=1)
	lcdE  = 5// set to high to process incoming data
	lcdD4 = 25//data pins used in 4 bot mode
	lcdD5 = 24
	lcdD6 = 23
	lcdD7 = 17

	//Define some device constants
	lcdWidth = 16 // Maximum characters per line
	lcdChr   = true
	lcdCmd   = false

	lcdLine1 = 0x80 // LCD RAM address for the 1st line
	lcdLine2 = 0xC0 // LCD RAM address for the 2nd line
)

func removeNlChars(str string) string {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	str, _, _ = transform.String(t, str)
	return str
}

//Lcd output
type Lcd struct {
	sync.Mutex //mutual exclusion using lock or unlock or defer
	lcdRS rpio.Pin// rpio is the adv gpio gpio for Pi and thus its used as datatype here
	lcdE  rpio.Pin
	lcdD4 rpio.Pin
	lcdD5 rpio.Pin
	lcdD6 rpio.Pin
	lcdD7 rpio.Pin

	line1  string
	line2  string
	active bool

	msg chan string //enables messages to be sent as channels in string form
	end chan bool
}

//NewLcd create and init new lcd output
func NewLcd() (l *Lcd) {
//exception handling on LCD
	if err := rpio.Open(); err != nil {
		panic(err)

	}

	l = &Lcd{
		lcdRS:  initPin(lcdRS), //pin creation and initializations
		lcdE:   initPin(lcdE),
		lcdD4:  initPin(lcdD4),
		lcdD5:  initPin(lcdD5),
		lcdD6:  initPin(lcdD6),
		lcdD7:  initPin(lcdD7),
		active: true,
		msg:    make(chan string),//making channel msg
		end:    make(chan bool), //making channel end
	}
	l.reset()

	go func() {
		for {
			select {
			case msg := <-l.msg: //sending LCD content into the already made channel
				l.display(msg)// displaying the same
			case _ = <-l.end:
				l.close()
				return
			}
		}
	}()
	return l
}//NewLcd()

//Display show some message
func (l *Lcd) Display(msg string) {
	l.msg <- msg //sending msg to l.msg
}//Display(str)

//Close LCD
func (l *Lcd) Close() {
	log.Printf("Lcd.Close")
	if l.active {
		l.end <- true
	}
}//Close()

func initPin(pin int) (p rpio.Pin) {
  p = rpio.Pin(pin)
  rpio.PinMode(p, rpio.Output)
  return
}//initPin(pin int) (p rpio.Pin)

func (l *Lcd) reset() {
  log.Printf("Lcd.reset()")
  //l.writeByte(0x33, lcdCmd) // 110011 Initialise
  l.write4Bits(0x3, lcdCmd) // 110011 Initialise
  time.Sleep(5 * time.Millisecond)
  //l.writeByte(0x32, lcdCmd) // 110010 Initialise
  l.write4Bits(0x3, lcdCmd) // 110010 Initialise
  time.Sleep(120 * time.Microsecond)
  //l.writeByte(0x30, lcdCmd) // 110000 Initialise
  l.write4Bits(0x3, lcdCmd) // 110010 Initialise
  time.Sleep(120 * time.Microsecond)

  l.write4Bits(0x2, lcdCmd) // 110010 Initialise
  time.Sleep(120 * time.Microsecond)

  l.writeByte(0x28, lcdCmd) // 101000 Data length, number of lines, font size on 4 bit mode
  l.writeByte(0x0C, lcdCmd) // 001100 Display On,Cursor Off, Blink Off
  l.writeByte(0x06, lcdCmd) // 000110 Cursor shift to right
  l.writeByte(0x01, lcdCmd) // 000001 Clear display screen
  time.Sleep(5 * time.Millisecond)
  log.Printf("Lcd.reset() finished")
}//reset()

func (l *Lcd) close() {
  l.Lock()
  defer l.Unlock()

  log.Printf("Lcd.close() active: %v", l.active)

  if !l.active {
  	return
  }

  l.writeByte(lcdLine1, lcdCmd)
  for i := 0; i < lcdWidth; i++ {
  	l.writeByte(' ', lcdChr)
  }
  l.writeByte(lcdLine2, lcdCmd)
  for i := 0; i < lcdWidth; i++ {
  	l.writeByte(' ', lcdChr)
  }
  time.Sleep(1 * time.Second)

  l.writeByte(0x01, lcdCmd) // 000001 Clear display
  l.writeByte(0x0C, lcdCmd) // 001000 Display on , cursor off

  l.lcdRS.Low()
  l.lcdE.Low()
  l.lcdD4.Low()
  l.lcdD5.Low()
  l.lcdD6.Low()
  l.lcdD7.Low()
  rpio.Close()

  l.active = false
  close(l.msg)
  close(l.end)
}//close()

// writeByte send byte to lcd
func (l *Lcd) writeByte(bits uint8, characterMode bool) {
  if characterMode {
  	l.lcdRS.High()
  } else {
  	l.lcdRS.Low()
  }

  //High bits
  if bits&0x10 == 0x10 {
  	l.lcdD4.High()
  } else {
  	l.lcdD4.Low()
  }
  if bits&0x20 == 0x20 {
  	l.lcdD5.High()
  } else {
  	l.lcdD5.Low()
  }
  if bits&0x40 == 0x40 {
  	l.lcdD6.High()
  } else {
  	l.lcdD6.Low()
  }
  if bits&0x80 == 0x80 {
  	l.lcdD7.High()
  } else {
  	l.lcdD7.Low()
  }

  //Toggle 'Enable' pin
  time.Sleep(toggle)
  l.lcdE.High()
  time.Sleep(toggle)
  l.lcdE.Low()
  time.Sleep(toggle)

  //Low bits
  if bits&0x01 == 0x01 {
  	l.lcdD4.High()
  } else {
  	l.lcdD4.Low()
  }
  if bits&0x02 == 0x02 {
  	l.lcdD5.High()
  } else {
  	l.lcdD5.Low()
  }
  if bits&0x04 == 0x04 {
  	l.lcdD6.High()
  } else {
  	l.lcdD6.Low()
  }
  if bits&0x08 == 0x08 {
  	l.lcdD7.High()
  } else {
  	l.lcdD7.Low()
  }
  //Toggle 'Enable' pin
  time.Sleep(toggle)
  l.lcdE.High()
  time.Sleep(toggle)
  l.lcdE.Low()

  time.Sleep(delay)
}//writeByte(bits uint8, characterMode bool)

//write4Bits send 4bits to lcd
func (l *Lcd) write4Bits(bits uint8, characterMode bool) {
  if characterMode {
  	l.lcdRS.High()
  } else {
  	l.lcdRS.Low()
  }

  //Low bits
  if bits&0x01 == 0x01 {
  	l.lcdD4.High()
  } else {
  	l.lcdD4.Low()
  }
  if bits&0x02 == 0x02 {
  	l.lcdD5.High()
  } else {
  	l.lcdD5.Low()
  }
  if bits&0x04 == 0x04 {
  	l.lcdD6.High()
  } else {
  	l.lcdD6.Low()
  }
  if bits&0x08 == 0x08 {
  	l.lcdD7.High()
  } else {
  	l.lcdD7.Low()
  }
  //Toggle 'Enable' pin
  time.Sleep(toggle)
  l.lcdE.High()
  time.Sleep(toggle)
  l.lcdE.Low()

  time.Sleep(delay)
}//write4Bits(bits uint8, characterMode bool)

func (l *Lcd) display(msg string) {
  l.Lock()
  defer l.Unlock()

  if !l.active {
  	return
  }

  log.Printf("Lcd.display('%#v')", msg)

  for line, m := range strings.Split(msg, "\n") {
  	if len(m) < lcdWidth {
  		m = m + strings.Repeat(" ", lcdWidth-len(m))
  	}

  	switch line {
  	case 0:
  		if l.line1 == m {
  			continue
  		}
  		l.line1 = m
  		l.writeByte(lcdLine1, lcdCmd)
  	case 1:
  		if l.line2 == m {
  			continue
  		}
  		l.line2 = m
  		l.writeByte(lcdLine2, lcdCmd)
  	default:
  		log.Printf("Lcd.display: to many lines %d: '%v'", line, m)
  		return
  	}

  	for i := 0; i < lcdWidth; i++ {
  		l.writeByte(byte(m[i]), lcdChr)
  	}
  }
}//display(msg string)

func geterror(e error){    //Error function

if e!=nil{

log.Println(e)

}

}


func main() {

	var disp Display
	log.Printf("main: starting lcd")
	disp = NewLcd()
	disp.Display("     \n     yeah Go!")

	var ( // necessary variable declarations

	 clientadd string

	 serveradd string

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


	snet.Init(client.IA, sciond.GetDefaultSCIONDPath(nil), daddr) //initialises scion network


	connectUDP, e = snet.DialSCION("udp4", client, server) // client connects to server through UDP

	geterror(e)

	//packetreceived := make([]byte, 3000) //Creating a  array of 3500 size for receiving

	//packetsent := make([]byte, 40)   //Creating a  array of 40 size for sending

	packetreceived := make([]string) //Creating a  array of 3500 size for receiving

	packetsent := make([]string)

	//seed := rand.NewSource(time.Now().UnixNano())




	for{
//Ts := time.Now() // sending the time now
//	_, e = connectUDP.Write(packetsent) //sending message to server
	//geterror(e)

	_, _, e = connectUDP.ReadFrom(packetreceived) // receiving message from server
	geterror(e)

  disp.Clear()//Add lcd clear

	disp.Display(string(receivePacketBuffer)
	time.Sleep(1 * time.Second)

}







defer func() {
  if e := recover(); e != nil {
    log.Printf("Recover: %v", e)
  }
  disp.Close()
  log.Printf("main.defer: all closed")
}()

} //main()
