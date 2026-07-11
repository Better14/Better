
package personext

import "person"

func (p person.Person) Hello() string {
	return "Hi, my name is " + p.Name
}
