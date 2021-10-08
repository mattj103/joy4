package mp4io

import (
	"github.com/mattj103/joy4/utils/bits/pio"
)

type MetaDataItem struct {
	Type   uint32
	Locale uint32
	Data   []byte
	AtomPos
}

const DATA = Tag(0x64617461)

func (self MetaDataItem) Tag() Tag {
	return DATA
}

func (self *MetaDataItem) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(DATA))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self *MetaDataItem) marshal(b []byte) (n int) {
	pio.PutU32BE(b[n:], self.Type)
	n += 4
	pio.PutU32BE(b[n:], self.Locale)
	n += 4
	if self.Data != nil {
		copy(b[n:], self.Data)
		n += len(self.Data)
	}
	return
}
func (self *MetaDataItem) Len() (n int) {
	n += 8
	n += 4
	n += 4
	n += len(self.Data)
	return
}
func (self *MetaDataItem) Unmarshal(b []byte, offset int) (n int, err error) {
	(&self.AtomPos).setPos(offset, len(b))
	self.Type = pio.U32BE(b[n:])
	n += 4
	self.Locale = pio.U32BE(b[n:])
	n += 4

	self.Data = b[n+4:]
	n += len(self.Data)
	return
}

func (self *MetaDataItem) Children() (r []Atom) {
	return
}


type MetaDataItemList struct {
	List map[string]*MetaDataItem
	AtomPos
}

const ILIST = Tag(0x696C7374)

func (self MetaDataItemList) Tag() Tag {
	return ILIST
}

func (self *MetaDataItemList) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(ILIST))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self *MetaDataItemList) marshal(b []byte) (n int) {
	if self.List != nil {
		for key, item := range self.List {
			pio.PutU32BE(b[n:], uint32(item.Len() + 8))
			n += 4
			copy(b[n:], []byte(key))
			n += 4
			n += item.Marshal(b[n:])
		}
	}
	return
}
func (self *MetaDataItemList) Len() (n int) {
	n += 8
	for _,item := range self.List {
		n += item.Len() + 8
    }
	return
}
func (self *MetaDataItemList) Unmarshal(b []byte, offset int) (n int, err error) {
	(&self.AtomPos).setPos(offset, len(b))
	self.List = make(map[string]*MetaDataItem)
	for n+8 < len(b) {
		size := int(pio.U32BE(b[n:]))
		key := string(b[n+4:n+8])
		n += 8
		if len(b) < n+size {
			err = parseErr("ListItemSizeInvalid", n+offset, err)
			return
		}
		item := new(MetaDataItem)
		if _, err = item.Unmarshal(b[n:n+size], offset+n); err != nil {
			err = parseErr("listitem", n+offset, err)
			return
		}
		self.List[key] = item
		n += size
	}
	return
}

func (self *MetaDataItemList) Children() (r []Atom) {
	return
}


type MetaData struct {
	Handler *HandlerRefer //This should be a metadata handler but works for my uses
	Items *MetaDataItemList
	Unknowns []Atom
	AtomPos
}

const META = Tag(0x6D657461)

func (self MetaData) Tag() Tag {
	return META
}

func (self *MetaData) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(META))
	n += self.marshal(b[12:]) + 8 + 4 //KEYS
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self *MetaData) marshal(b []byte) (n int) {
	// n += 4 //keys
	if self.Handler != nil {
		n += self.Handler.Marshal(b[n:])
	}
	if self.Items != nil {
		n += self.Items.Marshal(b[n:])
	}
	for _, atom := range self.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}
func (self *MetaData) Len() (n int) {
	n += 8
	n += 4 //keys
	if self.Handler != nil {
		n += self.Handler.Len()
	}
	if self.Items != nil {
		n += self.Items.Len()
	}
	for _, atom := range self.Unknowns {
		n += atom.Len()
	}
	return
}
func (self *MetaData) Unmarshal(b []byte, offset int) (n int, err error) {
	(&self.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case ILIST:
			{
				atom := &MetaDataItemList{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("ilst", n+offset, err)
					return
				}
				self.Items = atom
			}
		case HDLR:
			{
				atom := &HandlerRefer{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("hdlr", n+offset, err)
					return
				}
				self.Handler = atom
			}
		default:
			{
				atom := &Dummy{Tag_: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				self.Unknowns = append(self.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

func (self *MetaData) Children() (r []Atom) {
	if self.Handler != nil {
		r = append(r, self.Handler)
	}
	if self.Items != nil {
		r = append(r, self.Items)
	}
	r = append(r, self.Unknowns...)
	return
}

type UserData struct {
	List []Atom
	Unknowns []Atom
	AtomPos
}

const UDTA = Tag(0x75647461)

func (self UserData) Tag() Tag {
	return UDTA
}

func (self *UserData) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(UDTA))
	n += self.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (self *UserData) marshal(b []byte) (n int) {
	if self.List != nil {
		for _, atom := range self.List {
			n += atom.Marshal(b[n:])
		}
	}
	for _, atom := range self.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}
func (self *UserData) Len() (n int) {
	n += 8
	if self.List != nil {
		for _, atom := range self.List {
			n += atom.Len()
		}
	}
	for _, atom := range self.Unknowns {
		n += atom.Len()
	}
	return
}
func (self *UserData) Unmarshal(b []byte, offset int) (n int, err error) {
	(&self.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case META:
			{
				atom := &MetaData{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("meta", n+offset, err)
					return
				}
				self.List = append(self.List, atom)
			}
		default:
			{
				atom := &Dummy{Tag_: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				self.Unknowns = append(self.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

func (self *UserData) Children() (r []Atom) {
	r = append(r, self.List...)
	r = append(r, self.Unknowns...)
	return
}
