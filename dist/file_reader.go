package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/pion/mediadevices/pkg/wave"
	"io"
	"os"
	"time"
)

const (
	sampleRate = 48000
	channels   = 2
	sampleSize = 2
)

type AudioFile struct {
	rawReader      *os.File
	bufferedReader *bufio.Reader
	rawBuffer      []byte
	decoder        wave.Decoder
	ticker         *time.Ticker
}
func NewAudioFile(path string) (*AudioFile, error) {
	// Assume 48000 sample rate, mono channel, and S16LE interleaved
	latency := time.Millisecond * 120
	readFrequency := time.Second / latency
	readLen := sampleRate * channels * sampleSize / int(readFrequency)
	decoder, err := wave.NewDecoder(&wave.RawFormat{
		SampleSize:  sampleSize,
		IsFloat:     false,
		Interleaved: true,
	})

	fmt.Printf(`
Latency: %s
Read Frequency: %d Hz
Buffer Len: %d bytes
`, latency, readFrequency, readLen)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &AudioFile{
		rawReader:      f,
		bufferedReader: bufio.NewReader(f),
		rawBuffer:      make([]byte, readLen),
		decoder:        decoder,
		ticker:         time.NewTicker(latency),
	}, nil
}

func (file *AudioFile) Read() (chunk wave.Audio, release func(), err error) {
	_, err = io.ReadFull(file.bufferedReader, file.rawBuffer)
	if err != nil {
		// Keep looping the audio
		file.rawReader.Seek(0, 0)
		_, err = io.ReadFull(file.bufferedReader, file.rawBuffer)
		if err != nil {
			return
		}
	}

	chunk, err = file.decoder.Decode(binary.LittleEndian, file.rawBuffer, channels)
	if err != nil {
		return
	}

	int16Chunk := chunk.(*wave.Int16Interleaved)
	int16Chunk.Size.SamplingRate = sampleRate

	// Slow down reading so that it matches 48 KHz
	<-file.ticker.C
	return
}

func (file *AudioFile) Close() error {
	return file.rawReader.Close()
}

func (file *AudioFile) ID() string {
	return "raw-audio-from-file"
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}