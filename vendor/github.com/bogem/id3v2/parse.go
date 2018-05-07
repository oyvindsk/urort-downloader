// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"errors"
	"fmt"
	"io"
)

const frameHeaderSize = 10

var ErrUnsupportedVersion = errors.New("unsupported version of ID3 tag")
var errBlankFrame = errors.New("id or size of frame are blank")

type frameHeader struct {
	ID       string
	BodySize int64
}

// parse finds ID3v2 tag in rd and parses it to tag considering opts.
// If rd is smaller than expected, it returns ErrSmallHeaderSize.
func (tag *Tag) parse(rd io.Reader, opts Options) error {
	if rd == nil {
		return errors.New("rd is nil")
	}

	header, err := parseHeader(rd)
	if err == errNoTag || err == io.EOF {
		tag.init(rd, 0, 4)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error by parsing tag header: %v", err)
	}
	if header.Version < 3 {
		return ErrUnsupportedVersion
	}

	tag.init(rd, tagHeaderSize+header.FramesSize, header.Version)
	if !opts.Parse {
		return nil
	}
	return tag.parseFrames(opts)
}

func (tag *Tag) init(rd io.Reader, originalSize int64, version byte) {
	tag.DeleteAllFrames()

	tag.reader = rd
	tag.originalSize = originalSize
	tag.version = version
	tag.setDefaultEncoding(version)
}

func (tag *Tag) setDefaultEncoding(version byte) {
	if version == 4 {
		tag.defaultEncoding = EncodingUTF8
	} else {
		tag.defaultEncoding = EncodingISO
	}
}

func (tag *Tag) parseFrames(opts Options) error {
	framesSize := tag.originalSize - tagHeaderSize

	// Convert descriptions, specified by user in opts.ParseFrames, to IDs.
	parseIDs := make(map[string]bool, len(opts.ParseFrames))
	for _, description := range opts.ParseFrames {
		parseIDs[tag.CommonID(description)] = true
	}

	br := getBufReader(nil)
	defer putBufReader(br)

	buf := getByteSlice(32 * 1024)
	defer putByteSlice(buf)

	for framesSize > 0 {
		header, err := parseFrameHeader(buf, tag.reader)
		if err == io.EOF || err == errBlankFrame || err == ErrInvalidSizeFormat {
			break
		}
		if err != nil {
			return err
		}
		id := header.ID
		bodySize := header.BodySize

		framesSize -= frameHeaderSize + bodySize

		bodyRd := getLimitedReader(tag.reader, bodySize)
		defer putLimitedReader(bodyRd)

		if len(parseIDs) > 0 && !parseIDs[id] {
			if err := skipReaderBuf(bodyRd, buf); err != nil {
				return err
			}
			continue
		}

		br.Reset(bodyRd)
		frame, err := parseFrameBody(id, br)
		if err != nil && err != io.EOF {
			return err
		}

		tag.AddFrame(id, frame)

		if len(parseIDs) > 0 && !mustFrameBeInSequence(id) {
			delete(parseIDs, id)

			// If it was last ID in parseIDs, we don't need to parse
			// other frames, so end the parsing.
			if len(parseIDs) == 0 {
				break
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

func parseFrameHeader(buf []byte, rd io.Reader) (frameHeader, error) {
	var header frameHeader

	if len(buf) < frameHeaderSize {
		return header, errors.New("parseFrameHeader: buf is smaller than frame header size")
	}

	fhBuf := buf[:frameHeaderSize]
	if _, err := rd.Read(fhBuf); err != nil {
		return header, err
	}

	id := string(fhBuf[:4])
	bodySize, err := parseSize(fhBuf[4:8])
	if err != nil {
		return header, err
	}

	if id == "" || bodySize == 0 {
		return header, errBlankFrame
	}

	header.ID = id
	header.BodySize = bodySize
	return header, nil
}

// skipReaderBuf just reads rd until io.EOF.
func skipReaderBuf(rd io.Reader, buf []byte) error {
	for {
		_, err := rd.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func parseFrameBody(id string, br *bufReader) (Framer, error) {
	if id[0] == 'T' {
		return parseTextFrame(br)
	}

	if parseFunc, exists := parsers[id]; exists {
		return parseFunc(br)
	}

	return parseUnknownFrame(br)
}
