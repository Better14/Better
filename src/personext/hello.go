// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package personext

import "person"

func (p person.Person) Hello() string {
	return "Hi, my name is " + p.Name
}
