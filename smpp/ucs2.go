package smpp

import "github.com/linxGnu/gosmpp/data"

type ucs2f struct{}

func (ucs2f) Encode(str string) ([]byte, error) {
	return data.UCS2.Encode(str)
}

func (ucs2f) Decode(buf []byte) (string, error) {
	return data.UCS2.Decode(buf)
}

func (c ucs2f) ShouldSplit(text string, octetLimit uint) bool {
	encodedText, err := c.Encode(text)
	if err != nil {
		return true
	}
	return uint(len(encodedText)) > octetLimit
}

func (c ucs2f) EncodeSplit(text string, octetLimit uint) ([][]byte, error) {
	if octetLimit < 64 {
		octetLimit = 134
	}

	encodedText, err := c.Encode(text)
	if err != nil {
		return nil, err
	}

	var segCounter uint = 0

	allSeg := [][]byte{}
	var from uint = 0

	for _, r := range text {
		encodedRune, err := c.Encode(string(r))
		if err != nil {
			return nil, err
		}
		var size = uint(len(encodedRune))

		if segCounter+size <= octetLimit {
			segCounter += size
		} else {
			to := from + segCounter
			seg := encodedText[from:to]
			allSeg = append(allSeg, seg)
			from = to
			segCounter = size
		}
	}

	if segCounter > 0 {
		to := from + segCounter
		seg := encodedText[from:to]
		allSeg = append(allSeg, seg)
	}

	return allSeg, nil
}

func (ucs2f) DataCoding() byte { return data.UCS2Coding }
