## hrp run

run API test with go engine

### Synopsis

run yaml/json testcase files for API test

```
hrp run $path... [flags]
```

### Examples

```
  $ hrp run demo.json	# run specified json testcase file
  $ hrp run demo.yaml	# run specified yaml testcase file
  $ hrp run examples/	# run testcases in specified folder
```

### Options

```
      --case-timeout float32   set testcase timeout (seconds) (default 3600)
  -c, --continue-on-failure    continue running next step when failure occurs
  -g, --gen-html-report        generate html report
  -h, --help                   help for run
      --http-stat              turn on HTTP latency stat (DNSLookup, TCP Connection, etc.)
      --log-plugin             turn on plugin logging
      --log-requests-off       turn off request & response details logging
  -p, --proxy-url string       set proxy url
  -s, --save-tests             save tests summary
```

### Options inherited from parent commands

```
      --log-json           set log to json format (default colorized console)
  -l, --log-level string   set log level (default "INFO")
      --venv string        specify python3 venv path
```

### SEE ALSO

* [hrp](hrp.md)	 - All-in-One Testing Framework for API, UI and Performance

###### Auto generated by spf13/cobra on 24-Apr-2025
