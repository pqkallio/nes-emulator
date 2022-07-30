package rom

type NameTableMirroringType int

const (
	HorizontalOrMapperControlled NameTableMirroringType = iota
	Vertical
)

type ConsoleType int

const (
	NesFamicom ConsoleType = iota
	NintendoVs
	NintendoPlaychoice10
	ExtendedConsole
)

type ROM struct {
	prgROM                      []uint8
	chrROM                      []uint8
	miscellaneousROM            []uint8
	flags6                      uint8
	flags7                      uint8
	mapperFlags                 uint8
	prgRamFlags                 uint8
	chrRamFlags                 uint8
	timingFlags                 uint8
	systemTypeFlags             uint8
	miscellaneousRomFlags       uint8
	defaultExpansionDeviceFlags uint8
}

func (r *ROM) NameTableMirroringType() NameTableMirroringType {
	return NameTableMirroringType(r.flags6 & 0x01)
}

func (r *ROM) HasBattery() bool {
	return r.flags6&0x02 == 0x02
}

func (r *ROM) HasTrainer() bool {
	return r.flags6&0x04 == 0x04
}

func (r *ROM) HasHardWiredFourScreenMode() bool {
	return r.flags6&0x08 == 0x08
}

func (r *ROM) MapperNumber() uint16 {
	return uint16(r.flags6&0xf0)>>4 | uint16(r.flags7&0xf0) | uint16(r.mapperFlags&0x0f)<<8
}

func (r *ROM) ConsoleType() ConsoleType {
	return ConsoleType(r.flags7 & 0x03)
}

func (r *ROM) SubMapperNumber() uint8 {
	return (r.mapperFlags & 0xf0) >> 4
}
