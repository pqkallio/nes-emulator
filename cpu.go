package main

type Cpu struct {
	aReg   uint8
	xReg   uint8
	yReg   uint8
	sp     uint16
	pc     uint16
	status uint8
}
