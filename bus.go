package main

type Bus struct {
	addr uint16
	ram  *Ram
}

func (b *Bus) SetAddr(addr uint16) {
	b.addr = addr
	// TODO
}

func (b *Bus) WriteData(data uint8) {
	b.ram.Write(b.addr, data)
	// TODO
}

func (b *Bus) ReadData() uint8 {
	// TODO
	return 0
}
