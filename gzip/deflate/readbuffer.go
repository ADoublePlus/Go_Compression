package deflate

import (
	"io"
	"errors"
	"math/bits"
	"fmt"
)

type ReadBuffer struct {
	buf []uint8
	reader io.Reader
	numBytesLoaded int
	index int
	bitPosition int
}

func NewReadBuffer(reader io.Reader, bufferSize int) *ReadBuffer {
	rb := new(ReadBuffer)
	rb.buf = make([]uint8, bufferSize)
	rb.reader = reader
	rb.numBytesLoaded = 0
	rb.index = 0
	rb.bitPosition = 0 // Position of bit inside the byte at index 'rb.index',
	    			   // from the most-significant bit to the least-significant bit.
	return rb
}

func (rb *ReadBuffer) BitsLeftToRead() int {
	return rb.numBytesLoaded * 8 - (rb.index * 8 + rb.bitPosition)
}

func (rb *ReadBuffer) LoadMoreBytes() error {
	var numBytesRemaining = rb.numBytesLoaded - rb.index
	copy(rb.buf[0: numBytesRemaining], rb.buf[rb.index: rb.index + numBytesRemaining])
	rb.numBytesLoaded = numBytesRemaining
	n, err := rb.reader.Read(rb.buf[rb.numBytesLoaded: len(rb.buf)])
	rb.numBytesLoaded += n
	rb.index = 0

	if err != nil && err != io.EOF {
		return err
	} else {
		return nil
	}
}

/*func (rb *ReadBuffer) ReadBit() (bool, error) {
	if rb.index >= rb.numBytesLoaded {
		return false, errors.New("Index is out of bounds.")
	}

	var bit uint8 = (rb.buf[rb.index] << uint(rb.bitPosition)) & 0x80

	if rb.bitPosition < 7 {
		rb.bitPosition += 1
	} else {
		rb.index += 1
		rb.bitPosition = 0
	}

	if bit == 0 {
		return false, nil
	} else {
		return true, nil
	}
}*/

func (rb *ReadBuffer) ReadAlignedByte() (byte, error) {
	fmt.Printf("RB.ReadAlignedByte() index %d, bitPosition %d, numBytesLoaded %d\n", rb.index, rb.bitPosition, rb.numBytesLoaded)

	if rb.index >= rb.numBytesLoaded {
		return 0, errors.New("Index is out of bounds.")
	}

	rb.bitPosition = 0
	out := rb.buf[rb.index]
	rb.index += 1

	return out, nil
}

func (rb *ReadBuffer) ReadAlignedBytes(n int) ([]byte, int, error) {
	fmt.Printf("RB.ReadAlignedBytes() index %d, bitPosition %d, numBytesLoaded %d\n", rb.index, rb.bitPosition, rb.numBytesLoaded)

	if rb.index >= rb.numBytesLoaded {
		return nil, 0, errors.New("Index is out of bounds.")
	}

	rb.bitPosition = 0
	numBytesRemaining := rb.numBytesLoaded - rb.index

	if numBytesRemaining < n {
		n = numBytesRemaining
	}

	out := make([]byte, n)
	copy(out, rb.buf[rb.index: rb.index + n])
	rb.index += n

	return out, n, nil
}

func (rb *ReadBuffer) Peek() (uint64, error) {
	fmt.Printf("RB.Peek() index %d, bitPosition %d, numBytesLoaded %d\n", rb.index, rb.bitPosition, rb.numBytesLoaded)

	if rb.index >= rb.numBytesLoaded {
		return 0, errors.New("Index is out of bounds.")
	}

	indexEnd := rb.index + 9

	if indexEnd > rb.numBytesLoaded {
		indexEnd = rb.numBytesLoaded
	}

	return BytesToUint64WithBitReversal(rb.buf[rb.index: indexEnd], rb.bitPosition), nil
}

func (rb *ReadBuffer) Forward(n uint) error {
	bitIndex := rb.index * 8 + rb.bitPosition + int(n)
	fmt.Printf("RB.Forward() Before %d %d, After %d %d\n", rb.index, rb.bitPosition, int(bitIndex / 8), int(bitIndex % 8))

	if bitIndex > rb.numBytesLoaded * 8 {
		return errors.New("Number of bits to forward is too large: cannot forward to a position after the end of the buffer.")
	}

	rb.index = int(bitIndex / 8)
	rb.bitPosition = bitIndex % 8

	return nil
}

/*func (rb *ReadBuffer) Rewind(n uint) error {
	bitIndex := rb.index * 8 + rb.bitPosition - int(n)

	if bitIndex < 0 {
		return errors.New("Number of bits to rewind is too large: cannot rewind to a position before the start of the buffer.")
	}

	rb.index = int(bitIndex / 8)
	rb.bitPosition = bitIndex % 8

	return nil
}

type Prefix struct {
	data uint32
	index int
}

func NewPrefix() *Prefix {
	p := new(Prefix)
	p.Reset()
	return p
}

func (p *Prefix) Reset() {
	p.data = 0
	p.index = 0
}

func (p *Prefix) ReadBit(rb *ReadBuffer) error {
	if p.index >= 32 {
		return errors.New("Prefix is full, cannot add another bit to it.")
	}

	if bit, err := rb.ReadBit(); err != nil {
		return err
	} else if bit == true {
		p.data &= 1 << uint(31 - p.index)
	}

	p.index += 1

	return nil
}*/

func BytesToUint64(array []byte, bitOffset int) uint64 {
	var out uint64 = 0

	if len(array) < 8 {
		array = append(array, make([]byte, 8 - len(array))...)
	}

	if len(array) > 9 {
		panic("Invalid slice size")
	}

	if bitOffset > 8 {
		panic("Invalid bitOffset size")
	}

	out = (uint64(array[0]) << 56) |
		  (uint64(array[1]) << 48) |
		  (uint64(array[2]) << 40) |
		  (uint64(array[3]) << 32) |
		  (uint64(array[4]) << 24) |
		  (uint64(array[5]) << 16) |
		  (uint64(array[6]) <<  8) |
		  (uint64(array[7]))
	out = out << uint(bitOffset)

	if len(array) > 8 && bitOffset > 0 {
		out = out | (uint64(array[8]) >> uint(8 - bitOffset))
	}

	return out
}

func BytesToUint64WithBitReversal(array []byte, bitOffset int) uint64 {
	var out uint64 = 0

	if len(array) < 8 {
		array = append(array, make([]byte, 8 - len(array))...)
	}

	if len(array) > 9 {
		panic("Invalid slice size")
	}

	if bitOffset > 8 {
		panic("Invalid bitOffset size")
	}

	// 'array[0-7]' is [a7...a0][b7...b0] ... [h7...h0]
	// 'array[8]' is [i7...i0]

	out = (uint64(array[7]) << 56) |
		  (uint64(array[6]) << 48) |
		  (uint64(array[5]) << 40) |
		  (uint64(array[4]) << 32) |
		  (uint64(array[3]) << 24) |
		  (uint64(array[2]) << 16) |
		  (uint64(array[1]) <<  8) |
		  (uint64(array[0]))
	// 'out' is initialized as [h7...h0] ... [b7...b0][a7...a0]

	out = bits.Reverse64(out)
	// 'out' turned into [a0...a7][b0...b7] ... [h0...h7]

	if len(array) > 8 && bitOffset > 0 {
		out = (out << uint(bitOffset)) | (uint64(bits.Reverse8(array[8])) >> uint(8 - bitOffset))
	}

	// Example with bitOffset = 6
	// 'out' finally holds [a6...b5][b6...c5] ... [h6...i5]

	return out
}