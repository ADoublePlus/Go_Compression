package deflate

import (
	"testing"
	"strconv"
	//"fmt"
)

func TestGenerateCanonicalPrefixes(t *testing.T) {
	seq := []int{3, 3, 3, 3, 3, 2, 4, 4}
	GenerateCanonicalPrefixes(seq)

	/*fmt.Printf("------------------------------------------\n")

	for i := 0; i < len(seq); i++ {
		fmt.Printf("index:%d, length:%d code[%0*s]\n", i, seq[i], 32, strconv.FormatUint(uint64(codes[i]), 2))
	}*/
}

func TestGeneratorCanonicalPrefixesMode2(t *testing.T)  {
	litLenSequence := GenerateMode2LitLenSequence()
	codes := GenerateCanonicalPrefixes(litLenSequence)

	if len(codes) != 288 {
		t.Errorf("The generated code table has an invalid length, expected 288 and found %d", len(codes))
	}

	prefixes := []uint64{
		  0, 0x3000000000000000,
		143, 0xBF00000000000000,
		144, 0xC800000000000000,
		255, 0xFF80000000000000,
		256, 0x0000000000000000,
		279, 0x2E00000000000000,
		280, 0xC000000000000000,
		287, 0xC700000000000000,
	}

	for i := 0; i < len(prefixes); i += 2 {
		if codes[prefixes[i]] != prefixes[i + 1] {
			t.Errorf("Canonical codes general do not match the codes from the official Deflate specs.\nFor code '%d', expected [%0*s] and found [%0*s]\n", prefixes[i],
				64, strconv.FormatUint(uint64(prefixes[i + 1]), 2),
				64, strconv.FormatUint(uint64(codes[prefixes[i]]), 2))
		}
	}

	/*fmt.Printf("------------------------------------------\n")

	for i := 0; i < len(seq); i++ {
		fmt.Printf("index:%d, length:%d code[%0*s]\n", i, seq[i], 32, strconv.FormatUint(uint64(codes[i]), 2))
	}*/
}

func TestStuff(t *testing.T)  {
	litLenSequence := GenerateMode2LitLenSequence()
	distanceSequence := GenerateMode2DistanceSequence()
	NewTranslator(litLenSequence, distanceSequence)
	//tr := NewTranslator(litLenSequence, distanceSequence)
	//fmt.Printf("%#v\n", tr)
}