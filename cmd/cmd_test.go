package cmd

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
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

func Test_ExecuteCommandSimple(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)

	output, _ := executeCommand(rootCmd, "dump", "--path", "test.db", "--output", dir, "--all")

	if output != "" {
		t.Errorf("Got output %s", output)
	}

	outputpath := filepath.Join(dir, "Post.ndjson")
	outputfile, err := os.Stat(outputpath)
	if err != nil {
		log.Fatal(err)
	}

	content, err := ioutil.ReadFile(outputpath)
	if err != nil {
		log.Fatal(err)
	}
	checkStringContains(t, string(content), `Prisma is a database toolkit`)

	if outputfile.IsDir() {
		t.Errorf("Path should be a file")
	}
}
