// Package txt implements parsing and writing the UltraStar TXT file format.
// The implementation aims to be compatible with UltraStar and other compatible singing games.
//
// Unfortunately a formal specification of the UltraStar TXT format does not exist.
// The parser implementation of UltraStar has some weird edge cases that were intentionally left out in this package.
// The parser in this package should be able to parse most songs that are found in the wild.
//
// This package does not concern itself with the encoding of songs data.
// It expects all input to be UTF-8 encoded strings (or [io.Reader] reading UTF-8 bytes).
// However, many UltraStar songs found in the wild are actually in different encodings such as CP-1252.
// These are mostly compatible with UTF-8, but may produce wrong special characters.
// It is your responsibility to detect the encoding of the source data and convert it appropriately.
//
// There are some songs that make use of a special ENCODING tag.
// This package explicitly does not support this tag.
// If you want to process the tag you can either do some pre-processing or
// re-encode strings after they have been parsed.
//
// There are UltraStar TXTs known that use a UTF-8 byte order mark (BOM).
// A leading BOM has to be removed before this package can process the input.
// To remove the BOM you can use other packages
// such as [golang.org/x/text/encoding/unicode] or [github.com/dimchansky/utfbom].
package txt
