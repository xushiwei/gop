/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package html

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestText(t *testing.T) {
	cases := []struct {
		html string
		want string
	}{
		{`<p data-v-f3ebc54b="" class="param-description">Aspect ratio of the generated images
                            (width:height)</p>`, "Aspect ratio of the generated images (width:height)\n"},
	}
	for _, c := range cases {
		doc, e := html.Parse(strings.NewReader(c.html))
		if e != nil {
			t.Fatalf("html.Parse(%q) error: %v", c.html, e)
		}
		if got := textOf(doc); got != c.want {
			t.Errorf("Text(%q) = %q, want %q", c.html, got, c.want)
		}
	}
}
