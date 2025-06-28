package xmlutil

import (
	"encoding/xml"

	extxml "github.com/nbio/xml"
)

type UnmarshalDecoder = extxml.Decoder
type UnmarshalStartElement = extxml.StartElement

type MarshalEncoder = xml.Encoder
type MarshalStartElement = xml.StartElement

func Unmarshal(data []byte, v any) error {

	return extxml.Unmarshal(data, v)
}

const HeaderNoNewline = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"

func Marshal(v any) ([]byte, error) {
	res, err := xml.Marshal(v)
	if err != nil {
		return nil, err
	}
	return append([]byte(HeaderNoNewline), res...), nil
}

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	res, err := xml.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, err
	}
	return append([]byte(HeaderNoNewline+"\n"), res...), nil
}
