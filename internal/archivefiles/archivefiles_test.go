package archivefiles

import (
	"testing"
	"time"

	"github.com/goreleaser/goreleaser/internal/testlib"
	"github.com/goreleaser/goreleaser/internal/tmpl"
	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	ctx := context.New(config.Project{
		Env: []string{"OWNER=carlos"},
	})
	ctx.Git.CommitDate = now
	tmpl := tmpl.New(ctx)

	t.Run("templated info", func(t *testing.T) {
		result, err := Eval(tmpl, []config.File{
			{
				Source:      "./testdata/**/d.txt",
				Destination: "var/foobar/d.txt",
				Info: config.FileInfo{
					MTime: "{{.CommitDate}}",
					Owner: "{{ .Env.OWNER }}",
					Group: "{{ .Env.OWNER }}",
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, []config.File{
			{
				Source:      "testdata/a/b/c/d.txt",
				Destination: "var/foobar/d.txt/testdata/a/b/c/d.txt",
				Info: config.FileInfo{
					MTime:       now.UTC().Format(time.RFC3339),
					ParsedMTime: now.UTC(),
					Owner:       "carlos",
					Group:       "carlos",
				},
			},
		}, result)
	})

	t.Run("template info errors", func(t *testing.T) {
		t.Run("owner", func(t *testing.T) {
			_, err := Eval(tmpl, []config.File{{
				Source:      "./testdata/**/d.txt",
				Destination: "var/foobar/d.txt",
				Info: config.FileInfo{
					Owner: "{{ .Env.NOPE }}",
				},
			}})
			testlib.RequireTemplateError(t, err)
		})
		t.Run("group", func(t *testing.T) {
			_, err := Eval(tmpl, []config.File{{
				Source:      "./testdata/**/d.txt",
				Destination: "var/foobar/d.txt",
				Info: config.FileInfo{
					Group: "{{ .Env.NOPE }}",
				},
			}})
			testlib.RequireTemplateError(t, err)
		})
		t.Run("mtime", func(t *testing.T) {
			_, err := Eval(tmpl, []config.File{{
				Source:      "./testdata/**/d.txt",
				Destination: "var/foobar/d.txt",
				Info: config.FileInfo{
					MTime: "{{ .Env.NOPE }}",
				},
			}})
			testlib.RequireTemplateError(t, err)
		})
		t.Run("mtime format", func(t *testing.T) {
			_, err := Eval(tmpl, []config.File{{
				Source:      "./testdata/**/d.txt",
				Destination: "var/foobar/d.txt",
				Info: config.FileInfo{
					MTime: "2005-123-123",
				},
			}})
			require.Error(t, err)
		})
	})

	t.Run("single file", func(t *testing.T) {
		result, err := Eval(tmpl, []config.File{
			{
				Source:      "./testdata/**/d.txt",
				Destination: "var/foobar/d.txt",
			},
		})

		require.NoError(t, err)
		require.Equal(t, []config.File{
			{
				Source:      "testdata/a/b/c/d.txt",
				Destination: "var/foobar/d.txt/testdata/a/b/c/d.txt",
			},
		}, result)
	})

	t.Run("strip parent plays nicely with destination omitted", func(t *testing.T) {
		result, err := Eval(tmpl, []config.File{{Source: "./testdata/a/b", StripParent: true}})

		require.NoError(t, err)
		require.Equal(t, []config.File{
			{Source: "testdata/a/b/a.txt", Destination: "a.txt"},
			{Source: "testdata/a/b/c/d.txt", Destination: "d.txt"},
		}, result)
	})

	t.Run("strip parent plays nicely with destination as an empty string", func(t *testing.T) {
		result, err := Eval(tmpl, []config.File{{Source: "./testdata/a/b", Destination: "", StripParent: true}})

		require.NoError(t, err)
		require.Equal(t, []config.File{
			{Source: "testdata/a/b/a.txt", Destination: "a.txt"},
			{Source: "testdata/a/b/c/d.txt", Destination: "d.txt"},
		}, result)
	})

	t.Run("match multiple files within tree without destination", func(t *testing.T) {
		result, err := Eval(tmpl, []config.File{{Source: "./testdata/a"}})

		require.NoError(t, err)
		require.Equal(t, []config.File{
			{Source: "testdata/a/a.txt", Destination: "testdata/a/a.txt"},
			{Source: "testdata/a/b/a.txt", Destination: "testdata/a/b/a.txt"},
			{Source: "testdata/a/b/c/d.txt", Destination: "testdata/a/b/c/d.txt"},
		}, result)
	})

	t.Run("match multiple files within tree specific destination", func(t *testing.T) {
		result, err := Eval(tmpl, []config.File{
			{
				Source:      "./testdata/a",
				Destination: "usr/local/test",
				Info: config.FileInfo{
					Owner:       "carlos",
					Group:       "users",
					Mode:        0o755,
					ParsedMTime: now,
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, []config.File{
			{
				Source:      "testdata/a/a.txt",
				Destination: "usr/local/test/testdata/a/a.txt",
				Info: config.FileInfo{
					Owner:       "carlos",
					Group:       "users",
					Mode:        0o755,
					ParsedMTime: now,
				},
			},
			{
				Source:      "testdata/a/b/a.txt",
				Destination: "usr/local/test/testdata/a/b/a.txt",
				Info: config.FileInfo{
					Owner:       "carlos",
					Group:       "users",
					Mode:        0o755,
					ParsedMTime: now,
				},
			},
			{
				Source:      "testdata/a/b/c/d.txt",
				Destination: "usr/local/test/testdata/a/b/c/d.txt",
				Info: config.FileInfo{
					Owner:       "carlos",
					Group:       "users",
					Mode:        0o755,
					ParsedMTime: now,
				},
			},
		}, result)
	})

	t.Run("match multiple files within tree specific destination stripping parents", func(t *testing.T) {
		result, err := Eval(tmpl, []config.File{
			{
				Source:      "./testdata/a",
				Destination: "usr/local/test",
				StripParent: true,
				Info: config.FileInfo{
					Owner:       "carlos",
					Group:       "users",
					Mode:        0o755,
					ParsedMTime: now,
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, []config.File{
			{
				Source:      "testdata/a/a.txt",
				Destination: "usr/local/test/a.txt",
				Info: config.FileInfo{
					Owner:       "carlos",
					Group:       "users",
					Mode:        0o755,
					ParsedMTime: now,
				},
			},
			{
				Source:      "testdata/a/b/c/d.txt",
				Destination: "usr/local/test/d.txt",
				Info: config.FileInfo{
					Owner:       "carlos",
					Group:       "users",
					Mode:        0o755,
					ParsedMTime: now,
				},
			},
		}, result)
	})
}
