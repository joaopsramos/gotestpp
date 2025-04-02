package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func Test_process(t *testing.T) {
	color.NoColor = true

	originalStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = originalStdout
	})

	tests := []struct {
		name     string
		fileName string
		want     string
	}{
		{"success", "success.txt", successOutput},
		{"cached success", "success_cached.txt", successCachedOutput},
		{"build failed", "build_failed.txt", buildFailedOutput},
		{"skip", "skip.txt", skipOutput},
		{"fail", "fail.txt", failOutput},
		{"testify fail", "testify_fail.txt", testifyFailOutput},
		{"testify fail with message", "testify_fail_message.txt", testifyFailMessageOutput},
		{"fail and subtest fail", "fail_and_subtest_fail.txt", failAndSubTestFailOutput},
		{"fail with logs", "fail_with_logs.txt", failWithLogsOutput},
		{"fail with skip", "fail_with_skip.txt", failWithSkipOutput},
		{"multiple failures", "multiple_fails.txt", multipleFailsOutput},
		{"panic", "panic.txt", panicOutput},
		{"panic after assert", "panic_after_assert.txt", panicAfterAssertOutput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			parser := NewParser()
			renderer := NewRenderer()

			var wg sync.WaitGroup
			wg.Add(1)

			file, err := os.Open(filepath.Join("testdata", tt.fileName))
			a.NoError(err)
			defer file.Close()

			output := captureOutput(func() {
				process(&wg, file, parser, renderer)
			})

			a.Equal(tt.want, output)
		})
	}
}

var (
	successOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
ok	github.com/joaopsramos/fincon/internal/service	0.03s
ok	github.com/joaopsramos/fincon/internal/repository	0.03s
ok	github.com/joaopsramos/fincon/internal/api	0.58s

Finished in 0.63s
103 tests
`

	successCachedOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
ok	github.com/joaopsramos/fincon/internal/service	(cached)
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)

Finished in 0.00s
103 tests
`

	buildFailedOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
FAIL	github.com/joaopsramos/fincon/internal/service	[build failed]

Errors:
# github.com/joaopsramos/fincon/internal/service_test [github.com/joaopsramos/fincon/internal/service.test]
internal/service/expense_test.go:203:2: undefined: pan

Finished in 0.00s
94 tests
`

	skipOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
ok	github.com/joaopsramos/fincon/internal/service	(cached)
?	github.com/joaopsramos/fincon/internal/error	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)

--- SKIP TestExpenseService_Create (0.00s)
	expense_test.go:203

Finished in 0.00s
102 tests, 1 skipped
`

	failOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestExpenseService_Create (0.01s)
	expense_test.go:213: 2 should be equal to 1

Finished in 0.02s
103 tests, 1 failed
`

	testifyFailOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestPostgresExpense_GetSummary (0.01s)
--- FAIL TestPostgresExpense_GetSummary/should_handle_next_month_with_carried_over_excesses (0.00s)
	expense_test.go:97:
	Error:
		Not equal:
		expected: "\tone\ntwo\nthree\n\tfour"
		actual  : "\tone\nthree\ntwo\n\tfour"
		
		Diff:
		--- Expected
		+++ Actual
		@@ -1,4 +1,4 @@
			one
		+three
		 two
		-three
			four
	Error Trace:
		/home/joao/www/fincon/backend/internal/service/expense_test.go:97
		/home/joao/www/fincon/backend/internal/service/expense_test.go:199

Finished in 0.02s
100 tests, 2 failed
`

	testifyFailMessageOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestExpenseService_Create (0.01s)
	expense_test.go:213:
	Error:
		Not equal:
		expected: "one\ntwo\nthree\nfour"
		actual  : "one\nthree\nfour\ntwo"
		
		Diff:
		--- Expected
		+++ Actual
		@@ -1,4 +1,4 @@
		 one
		-two
		 three
		 four
		+two
	Messages:
		Some strange
			message
		here
	Error Trace:
		/home/joao/www/fincon/backend/internal/service/expense_test.go:213

Finished in 0.02s
103 tests, 1 failed
`

	failAndSubTestFailOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestExpenseService_Create (0.01s)
	expense_test.go:213: 2 should be equal to 1

--- FAIL TestExpenseService_Create/handle_float_precision_edge_cases (0.00s)
	expense_test.go:233: 1 should be equal to 2

Finished in 0.02s
103 tests, 2 failed
`

	failWithLogsOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
ok	github.com/joaopsramos/fincon/internal/api	(cached)
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestExpenseService_Create (0.01s)
--- FAIL TestExpenseService_Create/handle_float_precision_edge_cases (0.00s)
	expense_test.go:231: error to trigger log
log using fmt.Println
2025/03/31 14:59:44 log using log.Println

Finished in 0.02s
103 tests, 2 failed
`

	failWithSkipOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- SKIP TestExpenseService_Create (0.00s)
	expense_test.go:205

--- FAIL TestPostgresExpense_GetSummary (0.01s)
	expense_test.go:201:
	Error:
		Not equal:
		expected: 1
		actual  : 2
	Error Trace:
		/home/joao/www/fincon/backend/internal/service/expense_test.go:201

Finished in 0.02s
102 tests, 1 failed, 1 skipped
`

	multipleFailsOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestExpenseService_Create (0.01s)
	expense_test.go:213:
	Error:
		Not equal:
		expected: "one\ntwo\nthree\nfour"
		actual  : "one\nthree\nfour\ntwo"
		
		Diff:
		--- Expected
		+++ Actual
		@@ -1,4 +1,4 @@
		 one
		-two
		 three
		 four
		+two
	Messages:
		Some message
	Error Trace:
		/home/joao/www/fincon/backend/internal/service/expense_test.go:213
	expense_test.go:214:
	Error:
		Not equal:
		expected: "another"
		actual  : "assert"
		
		Diff:
		--- Expected
		+++ Actual
		@@ -1 +1 @@
		-another
		+assert
	Error Trace:
		/home/joao/www/fincon/backend/internal/service/expense_test.go:214
	expense_test.go:215:
	Error:
		Not equal:
		expected: "one"
		actual  : "more"
		
		Diff:
		--- Expected
		+++ Actual
		@@ -1 +1 @@
		-one
		+more
	Error Trace:
		/home/joao/www/fincon/backend/internal/service/expense_test.go:215

Finished in 0.02s
103 tests, 1 failed
`

	panicOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestPostgresExpense_GetSummary (0.01s)
panic: something went really wrong [recovered]
panic: something went really wrong
goroutine 20 [running]:
testing.tRunner.func1.2({0xb4e320, 0xdb4480})
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1635 +0x35e
panic({0xb4e320?, 0xdb4480?})
	/home/joao/.asdf/installs/golang/1.23.4/go/src/runtime/panic.go:785 +0x132
github.com/joaopsramos/fincon/internal/service_test.TestPostgresExpense_GetSummary(0xc0001c1d40)
	/home/joao/www/fincon/backend/internal/service/expense_test.go:33 +0x85
testing.tRunner(0xc0001c1d40, 0xce9d30)
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 1
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1743 +0x390

Finished in 0.01s
95 tests, 1 failed
`

	panicAfterAssertOutput = `?	github.com/joaopsramos/fincon/cmd/fincon	[no test files]
?	github.com/joaopsramos/fincon/cmd/migrate_db	[no test files]
?	github.com/joaopsramos/fincon/cmd/setup_db	[no test files]
?	github.com/joaopsramos/fincon/internal/config	[no test files]
?	github.com/joaopsramos/fincon/internal/domain	[no test files]
?	github.com/joaopsramos/fincon/internal/error	[no test files]
ok	github.com/joaopsramos/fincon/internal/repository	(cached)
ok	github.com/joaopsramos/fincon/internal/api	(cached)
?	github.com/joaopsramos/fincon/internal/testhelper	[no test files]
?	github.com/joaopsramos/fincon/internal/util	[no test files]
FAIL	github.com/joaopsramos/fincon/internal/service

--- FAIL TestPostgresExpense_GetSummary (0.01s)
	expense_test.go:33: Add(1,2) = 2;
        want 2000
panic: something went really wrong [recovered]
panic: something went really wrong
goroutine 20 [running]:
testing.tRunner.func1.2({0xb4e320, 0xdb44c0})
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1635 +0x35e
panic({0xb4e320?, 0xdb44c0?})
	/home/joao/.asdf/installs/golang/1.23.4/go/src/runtime/panic.go:785 +0x132
github.com/joaopsramos/fincon/internal/service_test.TestPostgresExpense_GetSummary(0xc0001c1d40)
	/home/joao/www/fincon/backend/internal/service/expense_test.go:34 +0xcd
testing.tRunner(0xc0001c1d40, 0xce9d50)
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 1
	/home/joao/.asdf/installs/golang/1.23.4/go/src/testing/testing.go:1743 +0x390

Finished in 0.01s
95 tests, 1 failed
`
)

func captureOutput(fn func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout = w
	fn()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
