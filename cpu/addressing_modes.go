package cpu

// accumulator accumulator addressing mode is represented with a one byte instruction,
// implying an operation on the accumulator.
func (c *Cpu) accumulator() uint8 {
	return 0
}

// immAddr the second byte of the instruction contains the operand.
func (c *Cpu) immAddr() uint8 {
	c.pc++
	c.absoluteAddr = c.pc

	return 0
}

// absAddr the second byte of the instruction contains the eight lower bits of
// the effective address and the third byte contains the eight higher bits.
func (c *Cpu) absAddr() uint8 {
	lo := uint16(c.bus.ReadData(c.pc))
	c.pc++
	hi := uint16(c.bus.ReadData(c.pc)) << 8
	c.pc++

	c.absoluteAddr = lo | hi

	return 0
}

// zeroPageAddr the second byte contains the eight lower bits of the effective
// address. The higher order bits are assumed to be zero.
func (c *Cpu) zeroPageAddr() uint8 {
	return c.indexedZeroPageAddr(0)
}

func (c *Cpu) xIndexedZeroPageAddr() uint8 {
	return c.indexedZeroPageAddr(c.xReg)
}

func (c *Cpu) yIndexedZeroPageAddr() uint8 {
	return c.indexedZeroPageAddr(c.yReg)
}

func (c *Cpu) indexedZeroPageAddr(idx uint8) uint8 {
	c.absoluteAddr = uint16(c.bus.ReadData(c.pc) + idx)
	c.pc++
	c.absoluteAddr &= 0x00FF

	return 0
}

func (c *Cpu) xIndexedAbsAddr() uint8 {
	return c.indexedAbsAddr(c.xReg)
}

func (c *Cpu) yIndexedAbsAddr() uint8 {
	return c.indexedAbsAddr(c.yReg)
}

func (c *Cpu) indexedAbsAddr(idx uint8) uint8 {
	lo := uint16(c.bus.ReadData(c.pc))
	c.pc++
	hi := uint16(c.bus.ReadData(c.pc)) << 8
	c.pc++

	c.absoluteAddr = (lo | hi) + uint16(idx)

	switch {
	case hi != (c.absoluteAddr & 0xff00):
		return 1
	default:
		return 0
	}
}

func (c *Cpu) impliedAddr() uint8 {
	return 0
}

func (c *Cpu) relAddr() uint8 {
	jump := c.bus.ReadData(c.pc)

	relativeAddr := uint16(jump)
	if jump&signMask != 0 {
		// If the 8-bit value is to be interpreted as a negative number,
		// the high byte should be all ones (2's complement).
		relativeAddr = 0xff00 & relativeAddr
	}

	c.relativeAddr = relativeAddr

	return 0
}

func (c *Cpu) indexedIndirectAddr() uint8 {
	ptr := uint16(c.bus.ReadData(c.pc))
	c.pc++

	loPtr := (ptr + uint16(c.xReg)) & 0x00ff
	hiPtr := (ptr + uint16(c.xReg) + 1) & 0x00ff

	addr := uint16(loPtr) | uint16(hiPtr)<<8

	c.absoluteAddr = addr

	return 0
}

func (c *Cpu) indirectIndexedAddr() uint8 {
	ptr := uint16(c.bus.ReadData(c.pc))
	c.pc++

	ptrLo := uint16(c.bus.ReadData(ptr & 0x00ff))
	ptrHi := uint16(c.bus.ReadData((ptr+1)&0x00ff)) << 8

	c.absoluteAddr = (ptrLo | ptrHi) + uint16(c.yReg)

	if c.absoluteAddr&0xff00 != ptrHi {
		return 1
	}

	return 0
}

func (c *Cpu) absIndirectAddr() uint8 {
	ptrLo := uint16(c.bus.ReadData(c.pc))
	c.pc++
	ptrHi := uint16(c.bus.ReadData(c.pc)) << 8
	c.pc++

	ptr := ptrLo | ptrHi

	var addrLo, addrHi uint16

	switch ptrLo {
	case 0x00ff:
		// Simulate a hardware bug when crossing page boundaries.
		addrLo = uint16(c.bus.ReadData(ptr))
		addrHi = uint16(c.bus.ReadData(ptr&0xff00)) << 8
	default:
		addrLo = uint16(c.bus.ReadData(ptr))
		addrHi = uint16(c.bus.ReadData(ptr+1)) << 8
	}

	c.absoluteAddr = addrHi | addrLo

	return 0
}
