/*
 Copyright 2022 The GoPlus Authors (goplus.org)
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package iox

import (
	"bufio"
	"io"
)

// ----------------------------------------------------------------------------

type LineIter struct {
	s *bufio.Scanner
}

func (it LineIter) Next() (line string, ok bool) {
	if ok = it.s.Scan(); ok {
		line = it.s.Text()
	}
	return
}

func EnumLines(r io.Reader) LineIter {
	scanner := bufio.NewScanner(r)
	return LineIter{scanner}
}

// ----------------------------------------------------------------------------

type LineReader struct {
	r io.Reader
}

func (p LineReader) Gop_Enum() LineIter {
	return EnumLines(p.r)
}

func Lines(r io.Reader) LineReader {
	return LineReader{r}
}

// ----------------------------------------------------------------------------
