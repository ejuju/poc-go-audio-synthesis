# Audio synthesis in Go

In order to produce sound (and eventually music) with our Go code, we need the following:
- An oscillator (like a sine wave that will produce sound at a certain frequency)
- A way to turn the continuous oscillator signal into audio frames
- A way to interpret these audio frames and play them on a speaker

Let's begin with our oscillator.
We'll use a simple sine wave for now, which can be defined as follows:

```go
func Sine(freq float64, x time.Duration) (y float64) {
    return math.Sin(x.Seconds() * 2 * math.Pi * freq)
}
```

NB: The frequency is in Hertz.

This function returns the value of our signal at the time "x" for the given frequency "freq" (in Hertz).

Now, if we want a speaker to play some sound using our sine function.
We must encode audio frames that our OS will pass on to the speakers.

To extract frames from a continuous signal, we need:
- A source signal
- A sample rate (how often we measure the value of our signal)
- Where to begin and end our sampling within the signal (offset and length)

But... what is a signal?
Simply, a signal is a function that returns a value (between -1 and 1) that fluctuates over time.
Like so:

```go
type Signal func(x time.Duration) (y float64)
```

Now let's refactor our previous sine function so that it returns a Signal function
that we can use later on to encode audio frames.

```go
func Sine(freq float64) Signal {
    return func(x time.Duration) (y float64) {
        return math.Sin(x.Seconds() * 2 * math.Pi * freq)
    }
}
```

Actually, we can even go one step further, let's say we want our oscillator's frequency to also change over time
(like a siren that goes from low to high frequency), then we can also make the input frequency argument to
be a signal:

```go
func Sine(freq Signal) Signal {
    return func(x time.Duration) (y float64) {
        return math.Sin(x.Seconds() * 2 * math.Pi * freq(x))
    }
}
```

But very often we will want our frequency to stay constant, so let's define a helper for that:

```go
func Constant(v float64) Signal {
    return func(x time.Duration) float64 { return v }
}
```

OK, so now we have a generic Signal type that we can use to:
- Generate sound at a given frequency with an oscillator
- Control other signals
- Sample audio frames

The next step is taking our continuous signal (which for now is just a mathemical function)
and getting audio frames:

```go
func Sample(s Signal, rate int, from, to time.Duration) (frames []float64) {
    step := float64(time.Second) / float64(rate)
	for i := float64(from); i < float64(from+to); i += step {
		val := s(time.Duration(i))
		frames = append(frames, val)
	}
	return frames
}
```

Where:
- `s` is our source signal.
- `rate` is the number of frames per second.
- `from` is where to begin measuring our signal.
- `to` is where to stop measuring our signal.

Cool, so now we went from a continuous signal to audio frames (= measurements) of the signal.
This allows computers to handle the input signal and play audio frames on a speaker.

But, how do we play audio frames?

There are several ways to go about this, but today we will be encoding frames
to a file (in a special format called PCM) and we can then use `ffplay` to
play this file on a speaker.

NB: We are not implementing the "playing" part (that's why we use ffplay),
we are only concerned with audio synthesis for now.

So, we said that we could encode our audio frames to PCM to play them with `ffplay`.
But what is PCM?

PCM is a very simple straightforward file format: audio frames are simply appended to the file,
one after another. The file doesn't have any header or anything else.
Since there's no way for the file to provide metadata, we will need to provide some information
to `ffplay`:
- The sample rate (ex: 44100 Hz)
- The frame encoding format (in our case: F64BE, since we encode each frame as a big-endian float64)

Let's write our PCM encoding function:
```go
func EncodePCM(frames []float64) (b []byte) {
    var buf [8]byte
    for _, pulse := range frames {
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(pulse))
		b = append(b, buf[:]...)
	}
	return b
}
```

We're ready to play our first sound!
Let's write a simple `main.go` file to create a 5-second-long audio PCM file that plays a sine oscillator at 440 Hz.

```go
func main() {
    signal := Sine(Constant(440))
    frames := Sample(signal, 44100, 0, 5*time.Second)
    os.Stdout.Write(EncodePCM(frames))
}
```

Let's run:
```shell
go run . > tmp/test.pcm && ffplay -f f64be -ar 44100 -autoexit -showmode 1 tmp/test.pcm
```

NB: We use `ffplay` to play raw PCM data here, but we could also encode our audio frames to a .WAV or .MP3 file.
