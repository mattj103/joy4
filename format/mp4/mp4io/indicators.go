package mp4io

import (
	"github.com/nareix/joy4/utils/bits/pio"
)

type TypeIndicator struct {
	Type byte
	AtomPos
}

func (self *TypeIndicator) Marshal(b []byte) (n int) {
	n += self.marshal(b)
	return
}
func (self *TypeIndicator) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], self.Type)
	n += 1
	return
}
func (self *TypeIndicator) Len() (n int) {
	n += 1
	return
}
func (self *TypeIndicator) Unmarshal(b []byte, offset int) (n int, err error) {
	(&self.AtomPos).setPos(offset, len(b))
	if len(b) < n+1 {
		err = parseErr("Type", n+offset, err)
		return
	}
	self.Type = pio.U8(b[n:])
	n += 1
	return
}
func (self *TypeIndicator) Children() (r []Atom) {
	return
}

type LocaleIndicator struct {
	Locale uint32
	AtomPos
}

func (self *LocaleIndicator) Marshal(b []byte) (n int) {
	n += self.marshal(b[0:])
	return
}
func (self *LocaleIndicator) marshal(b []byte) (n int) {
	pio.PutU24BE(b[n:], self.Locale)
	n += 3
	return
}
func (self *LocaleIndicator) Len() (n int) {
	n += 3
	return
}
func (self *LocaleIndicator) Unmarshal(b []byte, offset int) (n int, err error) {
	(&self.AtomPos).setPos(offset, len(b))
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	self.Locale = pio.U24BE(b[n:])
	n += 3
	return
}
func (self *LocaleIndicator) Children() (r []Atom) {
	return
}
