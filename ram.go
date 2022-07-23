package main

type Ram struct {
	internal     [0x800]uint16
	ppuRam       [0x8]uint16
	apuAndIO     [0x18]uint16
	apuAndIOTest [0x8]uint16
	cartridge    [0xbfe0]uint16
}

func NewRam() *Ram {
	return &Ram{}
}

func (r *Ram) Write(addr uint16, data uint8) {

}

func (r *Ram) Read(addr uint16) uint8 {
	return 0
}
