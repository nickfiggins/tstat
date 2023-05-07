package tstat

import "strings"

type Options struct {
	trimModule string
}

type ParseOpts func(*Options)

func TrimModule(name string) ParseOpts {
	return func(o *Options) {
		o.trimModule = name
	}
}

func (o Options) fileName(full string) string {
	if o.trimModule == "" {
		return strings.TrimPrefix(full, o.trimModule)
	}
	return full
}
