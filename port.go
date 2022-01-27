package debug_uart

import (
	"encoding/binary"
	"log"
	"time"

	"golang.org/x/sys/windows/registry"

	"github.com/tarm/serial"
)

type Uart struct {
	port             *serial.Port
	holdingRegisters []uint16
	Started          bool
	Stopped          bool
	state            byte
	error            byte
}

func Make() *Uart {

	uart := &Uart{}
	return uart
}

func (uart *Uart) Listen(name string) error {
	c := &serial.Config{Name: name, Baud: 115200}
	var err error
	uart.port, err = serial.OpenPort(c)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	return err
}

func (uart *Uart) Close() {
	uart.port.Close()
}

func (uart *Uart) Start() bool {

	if uart.Started {
		return true
	}

	var packet = []byte{'R', '1', '0', 255}

	n, err := uart.port.Write(packet)
	// log.Printf("%q", packet[:n])
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	packet = nil
	packet = []byte{'S', 't', '0', 255}
	n, err = uart.port.Write(packet)
	// log.Printf("%q", packet[:n])
	if err != nil || n == 0 {
		log.Fatal(err)
	}

	answer := make([]byte, 127)
	n1, err1 := uart.port.Read(answer)
	// log.Println(n)
	if err1 != nil || n1 == 0 {
		log.Fatal(err)
	}
	// time.Sleep(10 * time.Millisecond)
	// log.Println(n1)
	// log.Printf("%q", answer[:n1])

	uart.Started = (answer[3] == 66)
	uart.Stopped = false

	return uart.Started

}

func (uart *Uart) Stop() bool {
	if uart.Stopped {
		return true
	}

	var packet = []byte{'R', '0', 255}
	n, err := uart.port.Write(packet)
	// log.Printf("%q", packet[:n])
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	packet = nil
	packet = []byte{'S', 't', '0', 255}
	n, err = uart.port.Write(packet)
	// log.Printf("%q", packet[:n])
	if err != nil || n == 0 {
		log.Fatal(err)
	}

	answer := make([]byte, 127)
	n1, err1 := uart.port.Read(answer)
	// log.Println(n)
	if err1 != nil || n1 == 0 {
		log.Fatal(err)
	}
	// time.Sleep(10 * time.Millisecond)
	// log.Println(n1)
	// log.Printf("%q", answer[:n1])

	uart.Stopped = (answer[3] == 0)
	uart.Started = false

	return uart.Stopped
}

func (uart *Uart) GetData() []uint16 {

	// time.Sleep(30 * time.Millisecond)

	var packet = []byte{'C', 'V', '0', 255}
	uart.port.Write(packet)

	answer := make([]byte, 127)
	uart.port.Read(answer)
	if answer[0] == packet[0] && answer[1] == packet[1] {
		uart.holdingRegisters = bytesToUint16(answer[3:23])
	}
	return uart.holdingRegisters
}

func (uart *Uart) GetState() byte {

	var packet = []byte{'S', 't', '0', 255}
	uart.port.Write(packet)

	answer := make([]byte, 5)
	uart.port.Read(answer)
	if answer[0] == packet[0] && answer[1] == packet[1] {
		uart.state = answer[3]
	}
	return uart.state
}

func (uart *Uart) GetError() byte {
	var packet = []byte{'E', 'r', '0', 255}
	uart.port.Write(packet)

	answer := make([]byte, 5)
	uart.port.Read(answer)
	if answer[0] == packet[0] && answer[1] == packet[1] {
		uart.error = answer[3]
	}
	return uart.error
}

func bytesToUint16(bytes []byte) []uint16 {
	values := make([]uint16, len(bytes)/2)

	for i := range values {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2])
	}
	return values
}

func GetPort() []string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\\DEVICEMAP\\SERIALCOMM`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	ki, err := k.Stat()

	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("Subkey %d ValueCount %d\n", ki.SubKeyCount, ki.ValueCount)

	s, err := k.ReadValueNames(int(ki.ValueCount))
	if err != nil {
		log.Fatal(err)
	}
	kvalue := make([]string, ki.ValueCount)

	for i, test := range s {
		q, _, err := k.GetStringValue(test)
		if err != nil {
			log.Fatal(err)
		}
		kvalue[i] = q
	}

	return kvalue

	// fmt.Printf("%s \n", kvalue)
}
