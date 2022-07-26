package bus

import (
	"github.com/pqkallio/nes-emulator/ram"
)

type Bus struct {
	ram *ram.Ram
}

func (b *Bus) WriteData(addr uint16, data uint8) {
	b.ram.Write(addr, data)
}

func (b *Bus) ReadData(addr uint16) uint8 {
	return b.ram.Read(addr)
}
