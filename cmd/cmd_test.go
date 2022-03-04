package cmd

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)

	// Reset flags during tests since the flags persistent throughout cause issues
	defer resetFlags(rootCmd)
	return output, err
}

func executeCommandWithContext(ctx context.Context, root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.ExecuteContext(ctx)

	return buf.String(), err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func checkStringContains(t *testing.T, got, expected string) {
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}
}

// https://github.com/spf13/cobra/issues/770#issuecomment-627510928
func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Value.Type() == "stringSlice" {
			// XXX: unfortunately, flag.Value.Set() appends to original
			// slice, not resets it, so we retrieve pointer to the slice here
			// and set it to new empty slice manually
			value := reflect.ValueOf(flag.Value).Elem().FieldByName("value")
			ptr := (*[]string)(unsafe.Pointer(value.Pointer()))
			*ptr = make([]string, 0)
		}

		flag.Value.Set(flag.DefValue)
	})
	for _, cmd := range cmd.Commands() {
		resetFlags(cmd)
	}
}

func Test_ExecuteCommandHelp(t *testing.T) {
	output, err := executeCommand(rootCmd, "--help")
	checkStringContains(t, output, "Dump sqlite database metadata and table")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

}

func Test_ExecuteCommandRequiredFlags(t *testing.T) {
	output, _ := executeCommand(rootCmd, "dump")
	checkStringContains(t, output, `required flag(s) "output", "path" not set`)
}

func Test_ExecuteCommandNotExistentFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	output, _ := executeCommand(rootCmd, "dump", "--path", "nofile.db", "--output", dir)
	checkStringContains(t, output, `Path doesn't exist`)
}

func Test_ExecuteCommandAll(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)

	output, _ := executeCommand(rootCmd, "dump", "--path", "test.db", "--output", dir, "--all")

	if output != "" {
		t.Errorf("Got output %s", output)
	}

	outputPath := filepath.Join(dir, "Post.ndjson")
	outputFile, err := os.Stat(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	checkStringContains(t, string(content), `Prisma is a database toolkit`)

	if outputFile.IsDir() {
		t.Errorf("Path should be a file")
	}

	files := []string{"Post.ndjson", "Post1.ndjson", "Post.metadata.json", "Post1.metadata.json"}
	for _, file := range files {
		outputPath := filepath.Join(dir, file)
		_, err = os.Stat(outputPath)
		if err != nil {
			t.Errorf("%s should be present.", file)
		}
	}
}

func Test_ExecuteCommandNoAll(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)

	output, _ := executeCommand(rootCmd, "dump", "--path", "test.db", "--output", dir)

	checkStringContains(t, string(output), "You must pass --all or specify some tables")

	files := []string{"Post.ndjson", "Post1.ndjson", "Post.metadata.json", "Post1.metadata.json"}
	for _, file := range files {
		outputPath := filepath.Join(dir, file)
		_, err = os.Stat(outputPath)
		if err == nil {
			t.Errorf("%s should not be present.", file)
		}
	}
}

func Test_ExecuteCommandSingleTable(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)

	output, _ := executeCommand(rootCmd, "dump", "--path", "test.db", "--output", dir, "Post")

	if output != "" {
		t.Errorf("Got output %s", output)
	}

	outputPath := filepath.Join(dir, "Post.ndjson")
	_, err = os.Stat(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	checkStringContains(t, string(content), `Prisma is a database toolkit`)

	files := []string{"Post1.ndjson", "Post1.metadata.json"}
	for _, file := range files {
		outputPath := filepath.Join(dir, file)
		_, err = os.Stat(outputPath)
		if err == nil {
			t.Errorf("%s should not be present.", file)
		}
	}
}
