# go-ultrastar

[![Go Reference](https://pkg.go.dev/badge/codello.dev/ultrastar.svg)](https://pkg.go.dev/codello.dev/ultrastar)

This project provides the `ultrastar` package implementing the [UltraStar File Format](https://usdx.eu/format).
The package provides types for working with songs as well as a `Reader` and a `Writer` type for reading and writing the UltraStar file format.
This project supports all current versions of the file format.
Have a look at the [docs](https://pkg.go.dev/codello.dev/ultrastar).

## Installation

```shell
go get codello.dev/ultrastar
```

## Quick Start

```go
import "codello.dev/ultrastar"

// Parse song from a file
file, _ := os.Open("some/song.txt")
defer file.Close()
song, err := ultrastar.ReadSong(file)

// Do some transformations
song.Title = "Never Gonna Give You Up"
song.Voices[ultrastar.P1].ConvertToLeadingSpaces()

// Work with GAP, VIDEOGAP, etc. using native Go types
song.Gap = 2 * time.Second

// The ultrastar package provides convenient types for Pitches, Beats, BPM, ...
song.Voices[ultrastar.P1].Notes[2].Pitch = ultrastar.NamedPitch("F#2")

// Write song back to file
err := ultrastar.WriteSong(file, song, ultrastar.Version100)
```

Have a look at the [Docs](https://pkg.go.dev/codello.dev/ultrastar) to see everything you can do.
