package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", os.Args[0])

		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.PrintDefaults()

		fmt.Fprintln(flag.CommandLine.Output(), "\nEnvironment variables:")
		fmt.Fprintln(flag.CommandLine.Output(), "  BOTSU_TOKEN: Discord bot token")
		fmt.Fprintln(flag.CommandLine.Output(), "  BOTSU_CONNECTION_STRING: Database connection URL")
		fmt.Fprintln(flag.CommandLine.Output(), "  BOTSU_LOG_LEVEL: Log level")
		fmt.Fprintln(flag.CommandLine.Output(), "  BOTSU_AODB_PATH: Path to anime offline database")
		fmt.Fprintln(flag.CommandLine.Output(), "  BOTSU_ANIDB_DUMP_PATH: Path to anidb dump")
		fmt.Fprintln(flag.CommandLine.Output(), "  BOTSU_VNDB_DUMP_PATH: Path to vndb dump")
		fmt.Fprintln(flag.CommandLine.Output(), "  BOTSU_USE_MEMBERS_INTENT: Whether to use the members intent")

		fmt.Fprintln(flag.CommandLine.Output(), "\nConfig file:")
		fmt.Fprintln(flag.CommandLine.Output(), "The config file is a TOML file with the following structure:")
		fmt.Fprintln(flag.CommandLine.Output())
		printTOMLStructure(
			&prefixedWriter{w: flag.CommandLine.Output(), prefix: []byte("    ")},
			NewConfig(),
			"",
		)
	}
}

type prefixedWriter struct {
	w      io.Writer
	prefix []byte
}

func (pw *prefixedWriter) Write(p []byte) (n int, err error) {
	return pw.w.Write(append(pw.prefix, p...))
}

func printTOMLStructure(w io.Writer, v interface{}, topLevel string) {
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if topLevel != "" {
		fmt.Fprintf(w, "[%s]\n", topLevel)
	}

	structs := map[string]reflect.Type{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		name := f.Tag.Get("toml")

		if name == "-" {
			continue
		}

		switch f.Type.Kind() {
		case reflect.Struct:
			structs[name] = f.Type
		case reflect.Ptr:
			if f.Type.Elem().Kind() == reflect.Struct {
				structs[name] = f.Type.Elem()
			} else {
				fmt.Fprintf(w, "%s = *%s\n", name, f.Type.Elem().Kind())
			}
		default:
			fmt.Fprintf(w, "%s = %s\n", name, f.Type.Kind())
		}
	}

	fmt.Fprintln(w)

	for tag, s := range structs {
		printTOMLStructure(w, reflect.New(s).Interface(), tag)
	}
}
