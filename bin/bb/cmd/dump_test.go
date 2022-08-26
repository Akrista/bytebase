//go:build mysql
// +build mysql

package cmd

import (
	"fmt"
	"testing"

	// Embedded expected output.
	_ "embed"

	"github.com/stretchr/testify/require"

	"github.com/bytebase/bytebase/resources/mysql"
)

var (
	//go:embed testdata/expected/dump_test_TestDump
	_TestDump string
)

func TestDump(t *testing.T) {
	mysql, stop := mysql.SetupTestInstance(t, PortTestDump)
	defer stop()

	t.Log("Importing MySQL data...")
	err := mysql.Import("testdata/mysql_test_schema")
	require.NoError(t, err)

	tt := []testTable{
		{
			args: []string{
				"dump",
				"--dsn", fmt.Sprintf("mysql://root@localhost:%d/bytebase_test_todo", mysql.Port()),
			},
			expected: _TestDump,
		},
	}
	tableTest(t, tt)
}
