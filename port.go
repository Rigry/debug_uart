package uart

import (
	"log"

	"github.com/tarm/serial"
)

type Uart struct {
	port             *serial.Port
	holdingRegisters []uint16
}

func Make() *Uart {

	uart := &Uart{}
	return uart
}

func (uart *Uart) Listen() {
	c := &serial.Config{Name: "COM23", Baud: 9600}
	var err error
	uart.port, err = serial.OpenPort(c)

	if err != nil {
		log.Fatal(err)
	}
}

func (uart *Uart) Start() bool {

	var packet = []byte{'R', 'F', 1, 0, 'F', 'F'}

	n, err := uart.port.Write(packet)
	log.Println(packet[:n])
	if err != nil {
		log.Fatal(err)
	}

	packet = nil
	packet = []byte{'S', 't', 0, 'F', 'F'}
	n, err = uart.port.Write(packet)
	log.Println(packet[:n])
	if err != nil {
		log.Fatal(err)
	}

	answer := make([]byte, 127)
	n1, err1 := uart.port.Read(answer)
	// log.Println(n)
	if err1 != nil {
		log.Fatal(err)
	}
	log.Println(answer[:n1])

	return true
}
