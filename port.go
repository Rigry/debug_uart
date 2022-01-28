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
	switch_charger   bool
	state            byte
	error            byte
}

func Make() *Uart {

	uart := &Uart{}
	uart.holdingRegisters = make([]uint16, 10)
	return uart
}

func (uart *Uart) Listen(name string) error {
	c := &serial.Config{Name: name, Baud: 115200, ReadTimeout: time.Millisecond * 500}
	var err error
	uart.port, err = serial.OpenPort(c)

	return err
}

func (uart *Uart) Close() {
	uart.port.Close()
}

func (uart *Uart) Start() {

	var packet = []byte{'R', '1', 0, 255}
	uart.port.Write(packet)

	time.Sleep(10 * time.Millisecond)

	uart.Started = true
	uart.Stopped = false
}

func (uart *Uart) Stop() {

	var packet = []byte{'R', 0, 255}
	uart.port.Write(packet)

	time.Sleep(10 * time.Millisecond)

	uart.Stopped = true
	uart.Started = false
}

func (uart *Uart) GetData() []uint16 {

	answer := make([]byte, 25)
	var packet = []byte{'C', 'V', 0, 255}
	uart.port.Write(packet)
	time.Sleep(10 * time.Millisecond)
	uart.port.Read(answer)
	if answer[0] == packet[0] && answer[1] == 'v' && answer[23] == 255 {
		uart.holdingRegisters = bytesToUint16(answer[3:23])
	}
	return uart.holdingRegisters
}

func (uart *Uart) Get(packet []byte) {
	answer := make([]byte, 5)
	uart.port.Write(packet)
	time.Sleep(10 * time.Millisecond)
	uart.port.Read(answer)
	if answer[0] == 'S' && answer[1] == 't' && answer[4] == 255 {
		uart.state = answer[3]
	}
	if answer[0] == 'E' && answer[1] == 'r' && answer[4] == 255 {
		uart.error = answer[3]
	}
}

func (uart *Uart) GetState() byte {

	var packet = []byte{'S', 't', 0, 255}
	uart.Get(packet)
	return uart.state
}

func (uart *Uart) GetError() byte {

	var packet = []byte{'E', 'r', 0, 255}
	uart.Get(packet)
	return uart.error
}

func (uart *Uart) Charging() {
	if uart.switch_charger {
		var packet = []byte{'C', '0', 0, 255}
		uart.port.Write(packet)
		uart.switch_charger = false
	} else {
		var packet = []byte{'C', '1', 0, 255}
		uart.port.Write(packet)
		uart.switch_charger = true
	}
}

func (uart *Uart) Balance(n byte) {
	var packet = []byte{'d', n, 255}
	uart.port.Write(packet)
}

func (uart *Uart) Correction(n int, data uint16) {
	high, low := uint16ToBytes(data)
	var packet = []byte{'C', 'A', byte(n), high, low, 255}
	uart.port.Write(packet)
}

func bytesToUint16(bytes []byte) []uint16 {
	values := make([]uint16, len(bytes)/2)

	for i := range values {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2])
	}
	return values
}

func uint16ToBytes(data uint16) (high, low byte) {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes[:2], data)
	return bytes[0], bytes[1]
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
}
