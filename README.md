# go-ultrastar

This project provides multiple Go packages for working with [UltraStar](https://usdx.eu) songs. Have a look at the [Docs](https://pkg.go.dev/github.com/Karaoke-Manager/go-ultrastar).

## Packages

The main `ultrastar` package implements the main types for programmatically interacting with karaoke songs.

The `txt` subpackage implements a parser and a serializer for the UltraStar TXT format.

## Installation

```shell
go get github.com/Karaoke-Manager/go-ultrastar
```

## Quick Start

```go
package main

import (
  "github.com/Karaoke-Manager/go-ultrastar"
  "github.com/Karaoke-Manager/go-ultrastar/txt"
)

// Parse song from a file
file, _ := os.Open("some/song.txt")
defer file.Close()
song, err := txt.ReadSong(file)

// Do some transformations
song.Title = "Never Gonna Give You Up"
song.MusicP1.ConvertToLeadingSpaces()
// Work with GAP, VIDEOGAP, etc. using native Go types
song.Gap = 2 * time.Second
// The ultrastar package provides convenient types for Pitches, Beats, BPM, ...
song.MusicP1.Notes[2].Pitch = ultrastar.NamedPitch("F#2")

// Write song back to file
err = txt.WriteSong(file, song)
```

Have a look at the [Docs](https://pkg.go.dev/github.com/Karaoke-Manager/go-ultrastar) to see everything you can do.