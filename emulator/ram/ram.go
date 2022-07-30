package ram

var (
	ppuStart          uint16 = 0x2000
	apuAndIOStart     uint16 = 0x4000
	apuAndIOTestStart uint16 = 0x4018
	cartridgeStart    uint16 = 0x4020
)

type Ram struct {
	internal     [0x800]uint8
	ppuRam       [0x8]uint8
	apuAndIO     [0x18]uint8
	apuAndIOTest [0x8]uint8
	cartridge    [0xbfe0]uint8
}

func NewRam() *Ram {
	return &Ram{}
}

func (r *Ram) Write(addr uint16, data uint8) {
	switch {
	case addr < ppuStart:
		addr %= 0x800
		r.internal[addr] = data
	case addr < apuAndIOStart:
		addr -= ppuStart
		addr %= 0x8
		r.ppuRam[addr] = data
	case addr < apuAndIOTestStart:
		addr -= apuAndIOStart
		addr %= 0x18
		r.apuAndIO[addr] = data
	case addr < cartridgeStart:
		addr -= apuAndIOTestStart
		r.apuAndIOTest[addr] = data
	default:
		addr -= cartridgeStart
		r.cartridge[addr] = data
	}
}

func (r *Ram) Read(addr uint16) uint8 {
	switch {
	case addr < ppuStart:
		addr %= 0x800
		return r.internal[addr]
	case addr < apuAndIOStart:
		addr -= ppuStart
		addr %= 0x8
		return r.ppuRam[addr]
	case addr < apuAndIOTestStart:
		addr -= apuAndIOStart
		addr %= 0x18
		return r.apuAndIO[addr]
	case addr < cartridgeStart:
		addr -= apuAndIOTestStart
		return r.apuAndIOTest[addr]
	default:
		addr -= cartridgeStart
		return r.cartridge[addr]
	}
}
