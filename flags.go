package goapp

import "flag"

type appFlag struct {
	name        string
	description string
	value       any
}

func (app *App[_]) Flag(value any, name string, description string) {
	flag := appFlag{
		name:        name,
		description: description,
		value:       value,
	}

	app.f.flags = append(app.f.flags, flag)
}

func (app *App[_]) registerFlags() {
	for _, f := range app.f.flags {
		switch v := f.value.(type) {
		case *string:
			flag.StringVar(v, f.name, "", "")
		case *int:
			flag.IntVar(v, f.name, 0, "")
		case *bool:
			flag.BoolVar(v, f.name, false, "")
		}
	}
}

func (app *App[_]) getFlagHelp() string {
	s := ""

	for _, f := range app.f.flags {
		prefix := "-"
		if len(f.name) > 1 {
			prefix = "--"
		}

		s += "\n\t" + prefix + f.name + ": " + f.description
	}

	return s
}
