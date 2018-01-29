package flags

import (
	"fmt"
	"html/template"
)

type TemplateFlags []string

func (t *TemplateFlags) String() string {
	return fmt.Sprintf("%v", *t)
}

func (t *TemplateFlags) Set(value string) error {

	*t = append(*t, value)
	return nil
}

func (t *TemplateFlags) Parse() error {

     if len(*t) == 0 {
     	return nil
     }

     _, err := template.ParseFiles(*t...)
     return err
}
