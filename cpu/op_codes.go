package cpu

import "reflect"

func (c *Cpu) adc() uint8 {
	c.fetchData()
	data := c.fetchedData

	carry := uint16(0)
	if c.getFlag(carryFlag) {
		carry = 1
	}

	result := uint16(c.aReg) + uint16(data) + carry

	// Set status flags
	c.setFlag(carryFlag, result > 0xff)                      // if the result didn't fit in one byte, set the carry flag
	c.setFlag(zeroFlag, result&0x00ff == 0)                  // if the result is zero, set the zero flag
	c.setFlag(negativeFlag, result&uint16(signMask) != 0x00) // if the leftmost bit in the result is set, set the negative flag
	// if the accumulator and operand are negative and the result is positive,
	// or if the accumulator and operand are positive and the result is negative, set the overflow flag
	c.setFlag(
		overflowFlag,
		(c.aReg&signMask != 0 && data&signMask != 0 && result&uint16(signMask) == 0) ||
			(c.aReg&signMask == 0 && data&signMask == 0 && result&uint16(signMask) != 0),
	)

	c.aReg = uint8(result & 0x00ff)

	return 1
}

func (c *Cpu) and() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.aReg = c.aReg & data
	c.setFlag(zeroFlag, c.aReg == 0x00)
	c.setFlag(negativeFlag, c.aReg&signMask != 0x00)

	return 1
}

func (c *Cpu) asl() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(carryFlag, data&signMask != 0)
	data <<= 1
	c.setFlag(zeroFlag, data == 0)
	c.setFlag(negativeFlag, data&signMask != 0)

	c.writeToMem(data)

	return 0
}

func (c *Cpu) bcc() uint8 {
	return c.branchOn(!c.getFlag(carryFlag))
}

func (c *Cpu) bcs() uint8 {
	return c.branchOn(c.getFlag(carryFlag))
}

func (c *Cpu) beq() uint8 {
	return c.branchOn(c.getFlag(zeroFlag))
}

func (c *Cpu) bit() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(zeroFlag, c.aReg&data == 0)
	c.setFlag(negativeFlag, data&uint8(negativeFlag) != 0)
	c.setFlag(overflowFlag, data&uint8(overflowFlag) != 0)

	return 0
}

func (c *Cpu) bmi() uint8 {
	return c.branchOn(c.getFlag(negativeFlag))
}

func (c *Cpu) bne() uint8 {
	return c.branchOn(!c.getFlag(zeroFlag))
}

func (c *Cpu) bpl() uint8 {
	return c.branchOn(!c.getFlag(negativeFlag))
}

func (c *Cpu) brk() uint8 {
	c.setFlag(breakFlag, true)
	c.pc++
	return 0
}

func (c *Cpu) bvc() uint8 {
	return c.branchOn(!c.getFlag(overflowFlag))
}

func (c *Cpu) bvs() uint8 {
	return c.branchOn(c.getFlag(overflowFlag))
}

func (c *Cpu) clc() uint8 {
	c.setFlag(carryFlag, false)
	return 0
}

func (c *Cpu) cld() uint8 {
	c.setFlag(decimalModeFlag, false)
	return 0
}

func (c *Cpu) cli() uint8 {
	c.setFlag(disableInterruptsFlag, false)
	return 0
}

func (c *Cpu) clv() uint8 {
	c.setFlag(overflowFlag, false)
	return 0
}

func (c *Cpu) cmp() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(carryFlag, c.aReg >= data)
	c.setFlag(zeroFlag, c.aReg == data)
	c.setFlag(negativeFlag, c.aReg&signMask != 0)

	return 0
}

func (c *Cpu) cpx() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(carryFlag, c.xReg >= data)
	c.setFlag(zeroFlag, c.xReg == data)
	c.setFlag(negativeFlag, c.xReg&signMask != 0)

	return 0
}

func (c *Cpu) cpy() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(carryFlag, c.yReg >= data)
	c.setFlag(zeroFlag, c.yReg == data)
	c.setFlag(negativeFlag, c.yReg&signMask != 0)

	return 0
}

func (c *Cpu) dec() uint8 {
	c.fetchData()
	data := c.fetchedData

	result := data - 1

	c.setFlag(negativeFlag, result&signMask != 0)
	c.setFlag(zeroFlag, result == 0)

	c.writeToMem(result)

	return 0
}

func (c *Cpu) dex() uint8 {
	result := c.xReg - 1

	c.setFlag(negativeFlag, result&signMask != 0)
	c.setFlag(zeroFlag, result == 0)

	c.xReg = result

	return 0
}

func (c *Cpu) dey() uint8 {
	result := c.yReg - 1

	c.setFlag(negativeFlag, result&signMask != 0)
	c.setFlag(zeroFlag, result == 0)

	c.yReg = result

	return 0
}

func (c *Cpu) eor() uint8 {
	c.fetchData()
	data := c.fetchedData

	result := c.aReg ^ data

	c.setFlag(negativeFlag, result&signMask != 0)
	c.setFlag(zeroFlag, result == 0)

	c.aReg = result

	return 1
}

func (c *Cpu) inc() uint8 {
	c.fetchData()
	data := c.fetchedData

	result := data + 1

	c.setFlag(negativeFlag, result&signMask != 0)
	c.setFlag(zeroFlag, result == 0)

	c.writeToMem(result)

	return 0
}

func (c *Cpu) inx() uint8 {
	result := c.xReg + 1

	c.setFlag(negativeFlag, result&signMask != 0)
	c.setFlag(zeroFlag, result == 0)

	c.xReg = result

	return 0
}

func (c *Cpu) iny() uint8 {
	result := c.yReg + 1

	c.setFlag(negativeFlag, result&signMask != 0)
	c.setFlag(zeroFlag, result == 0)

	c.yReg = result

	return 0
}

func (c *Cpu) jmp() uint8 {
	c.pc = c.absoluteAddr

	return 0
}

func (c *Cpu) jsr() uint8 {
	c.bus.WriteData(stackBase|uint16(c.sp), uint8(c.pc>>8))
	c.sp--
	c.bus.WriteData(stackBase|uint16(c.sp), uint8(c.pc))
	c.sp--

	c.pc = c.absoluteAddr

	return 0
}

func (c *Cpu) lda() uint8 {
	c.fetchData()
	c.aReg = c.fetchedData

	return 0
}

func (c *Cpu) ldx() uint8 {
	c.fetchData()
	c.xReg = c.fetchedData

	return 0
}

func (c *Cpu) ldy() uint8 {
	c.fetchData()
	c.yReg = c.fetchedData

	return 0
}

func (c *Cpu) lsr() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.setFlag(carryFlag, data&1 != 0)
	data >>= 1
	c.setFlag(zeroFlag, data == 0)
	c.setFlag(negativeFlag, false)

	c.writeToMem(data)

	return 0
}

func (c *Cpu) nop() uint8 {
	return 0
}

func (c *Cpu) ora() uint8 {
	c.fetchData()
	data := c.fetchedData

	c.aReg = c.aReg | data

	return 0
}

func (c *Cpu) pha() uint8 {
	c.bus.WriteData(stackBase+uint16(c.sp), c.aReg)
	c.sp--
	return 0
}

func (c *Cpu) php() uint8 {
	c.bus.WriteData(stackBase+uint16(c.sp), c.status)
	c.sp--
	return 0
}

func (c *Cpu) pla() uint8 {
	c.sp++
	c.aReg = c.bus.ReadData(stackBase + uint16(c.sp))

	c.setFlag(zeroFlag, c.aReg == 0x00)
	c.setFlag(negativeFlag, c.aReg&signMask != 0x00)

	return 0
}

func (c *Cpu) plp() uint8 {
	c.sp++
	c.status = c.bus.ReadData(stackBase + uint16(c.sp))

	return 0
}

func (c *Cpu) rol() uint8 {
	c.fetchData()
	data := c.fetchedData

	carry := 0
	if data&signMask != 0 {
		carry = 1
	}

	data <<= 1
	if c.getFlag(carryFlag) {
		data |= 1
	}

	c.setFlag(carryFlag, carry != 0)
	c.setFlag(zeroFlag, data == 0)
	c.setFlag(negativeFlag, data&signMask != 0)

	c.writeToMem(data)

	return 0
}

func (c *Cpu) ror() uint8 {
	c.fetchData()
	data := c.fetchedData

	carry := 0
	if data&0x01 != 0 {
		carry = 1
	}

	data >>= 1
	if c.getFlag(carryFlag) {
		data |= signMask
	}

	c.setFlag(carryFlag, carry != 0)
	c.setFlag(zeroFlag, data == 0)
	c.setFlag(negativeFlag, data&signMask != 0)

	c.writeToMem(data)

	return 0
}

func (c *Cpu) rti() uint8 {
	c.sp++
	status := c.bus.ReadData(stackBase | uint16(c.sp))
	status &= ^uint8(breakFlag)
	status &= ^uint8(unusedFlag)
	c.sp++
	addrHi := uint16(c.bus.ReadData(stackBase | uint16(c.sp)))
	c.sp++
	addrLo := uint16(c.bus.ReadData(stackBase|uint16(c.sp))) << 8

	addr := addrHi | addrLo

	c.pc = addr
	c.status = status

	return 0
}

func (c *Cpu) rts() uint8 {
	c.sp++
	addrHi := uint16(c.bus.ReadData(stackBase | uint16(c.sp)))
	c.sp++
	addrLo := uint16(c.bus.ReadData(stackBase|uint16(c.sp))) << 8

	addr := addrHi | addrLo

	c.pc = addr + 1

	return 0
}

func (c *Cpu) sbc() uint8 {
	// The substraction for unsigned numbers can be achieved with the following formula:
	// accum + ~memory + (1 - carry)
	c.fetchData()
	m := ^uint16(c.fetchedData)

	carry := uint16(1)
	if c.getFlag(carryFlag) {
		// When substracting, if the carry flag has been set, it means that a borrow has occured.
		carry = 0
	}

	result := uint16(c.aReg) + m + carry

	// Set status flags
	c.setFlag(carryFlag, result > 0xff)                      // if the result didn't fit in one byte, set the carry flag
	c.setFlag(zeroFlag, result&0x00ff == 0)                  // if the result is zero, set the zero flag
	c.setFlag(negativeFlag, result&uint16(signMask) != 0x00) // if the leftmost bit in the result is set, set the negative flag
	// if the accumulator and operand are negative and the result is positive,
	// or if the accumulator and operand are positive and the result is negative, set the overflow flag
	c.setFlag(
		overflowFlag,
		(c.aReg&signMask != 0 && m&uint16(signMask) != 0 && result&uint16(signMask) == 0) ||
			(c.aReg&signMask == 0 && m&uint16(signMask) == 0 && result&uint16(signMask) != 0),
	)

	c.aReg = uint8(result & 0x00ff)

	return 1
}

func (c *Cpu) sec() uint8 {
	c.setFlag(carryFlag, true)
	return 0
}

func (c *Cpu) sed() uint8 {
	c.setFlag(decimalModeFlag, true)
	return 0
}

func (c *Cpu) sei() uint8 {
	c.setFlag(disableInterruptsFlag, true)
	return 0
}

func (c *Cpu) sta() uint8 {
	c.writeToMem(c.aReg)
	return 0
}

func (c *Cpu) stx() uint8 {
	c.writeToMem(c.xReg)
	return 0
}

func (c *Cpu) sty() uint8 {
	c.writeToMem(c.yReg)
	return 0
}

func (c *Cpu) tsx() uint8 {
	c.xReg = c.sp
	return 0
}

func (c *Cpu) txs() uint8 {
	c.sp = c.xReg
	return 0
}

func (c *Cpu) tax() uint8 {
	c.xReg = c.aReg

	c.setFlag(zeroFlag, c.xReg == 0)
	c.setFlag(negativeFlag, c.xReg&signMask != 0)

	return 0
}

func (c *Cpu) tay() uint8 {
	c.yReg = c.aReg

	c.setFlag(zeroFlag, c.yReg == 0)
	c.setFlag(negativeFlag, c.yReg&signMask != 0)

	return 0
}

func (c *Cpu) txa() uint8 {
	c.aReg = c.xReg

	c.setFlag(zeroFlag, c.aReg == 0)
	c.setFlag(negativeFlag, c.aReg&signMask != 0)

	return 0
}

func (c *Cpu) tya() uint8 {
	c.aReg = c.yReg

	c.setFlag(zeroFlag, c.aReg == 0)
	c.setFlag(negativeFlag, c.aReg&signMask != 0)

	return 0
}

func (c *Cpu) illegal() uint8 {
	return 0
}

func (c *Cpu) branchOn(cond bool) uint8 {
	if !cond {
		return 0
	}

	c.nCycles++

	c.absoluteAddr = c.pc + c.relativeAddr

	if c.absoluteAddr&0xff00 != c.pc&0xff00 {
		c.nCycles++
	}

	c.pc = c.absoluteAddr

	return 0
}

func (c *Cpu) writeToMem(val uint8) {
	switch reflect.ValueOf(c.opCodeLookup[c.opCode].addr).Pointer() {
	case reflect.ValueOf(c.accumulator).Pointer():
		c.aReg = val
	default:
		c.bus.WriteData(c.absoluteAddr, val)
	}
}
