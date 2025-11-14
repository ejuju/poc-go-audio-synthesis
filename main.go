package main

import (
	"encoding/binary"
	"math"
	"os"
	"time"
)

func main() {
	signal := Sine(Constant(440))
	frames := Sample(signal, 44100, 0, 5*time.Second)
	os.Stdout.Write(EncodePCM(frames))
}

type Signal func(x time.Duration) (y float64)

func Constant(v float64) Signal {
	return func(x time.Duration) float64 { return v }
}

func Sine(freq Signal) Signal {
	return func(x time.Duration) (y float64) {
		return math.Sin(x.Seconds() * 2 * math.Pi * freq(x))
	}
}

func Sample(s Signal, rate int, from, to time.Duration) (frames []float64) {
	step := float64(time.Second) / float64(rate)
	for i := float64(from); i < float64(from+to); i += step {
		val := s(time.Duration(i))
		frames = append(frames, val)
	}
	return frames
}

func EncodePCM(frames []float64) (b []byte) {
	var buf [8]byte
	for _, pulse := range frames {
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(pulse))
		b = append(b, buf[:]...)
	}
	return b
}
