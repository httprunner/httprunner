# locusts

## Installation

```shell script
$ pip install locustio
```

## Usage

```shell script
$ locusts -f xxx.yml
```

```shell script
$ locusts -f xxx.yml --processes
```

```shell script
$ python3 -m httprunner.ext.locusts -h

Usage: locust [options] [LocustClass [LocustClass2 ... ]]

Options:
  -h, --help            show this help message and exit
  -H HOST, --host=HOST  Host to load test in the following format:
                        http://10.21.32.33
  --web-host=WEB_HOST   Host to bind the web interface to. Defaults to '' (all
                        interfaces)
  -P PORT, --port=PORT, --web-port=PORT
                        Port on which to run web host
  -f LOCUSTFILE, --locustfile=LOCUSTFILE
                        Python module file to import, e.g. '../other.py'.
                        Default: locustfile
  --csv=CSVFILEBASE, --csv-base-name=CSVFILEBASE
                        Store current request stats to files in CSV format.
  --master              Set locust to run in distributed mode with this
                        process as master
  --slave               Set locust to run in distributed mode with this
                        process as slave
  --master-host=MASTER_HOST
                        Host or IP address of locust master for distributed
                        load testing. Only used when running with --slave.
                        Defaults to 127.0.0.1.
  --master-port=MASTER_PORT
                        The port to connect to that is used by the locust
                        master for distributed load testing. Only used when
                        running with --slave. Defaults to 5557. Note that
                        slaves will also connect to the master node on this
                        port + 1.
  --master-bind-host=MASTER_BIND_HOST
                        Interfaces (hostname, ip) that locust master should
                        bind to. Only used when running with --master.
                        Defaults to * (all available interfaces).
  --master-bind-port=MASTER_BIND_PORT
                        Port that locust master should bind to. Only used when
                        running with --master. Defaults to 5557. Note that
                        Locust will also use this port + 1, so by default the
                        master node will bind to 5557 and 5558.
  --heartbeat-liveness=HEARTBEAT_LIVENESS
                        set number of seconds before failed heartbeat from
                        slave
  --heartbeat-interval=HEARTBEAT_INTERVAL
                        set number of seconds delay between slave heartbeats
                        to master
  --expect-slaves=EXPECT_SLAVES
                        How many slaves master should expect to connect before
                        starting the test (only when --no-web used).
  --no-web              Disable the web interface, and instead start running
                        the test immediately. Requires -c and -r to be
                        specified.
  -c NUM_CLIENTS, --clients=NUM_CLIENTS
                        Number of concurrent Locust users. Only used together
                        with --no-web
  -r HATCH_RATE, --hatch-rate=HATCH_RATE
                        The rate per second in which clients are spawned. Only
                        used together with --no-web
  -t RUN_TIME, --run-time=RUN_TIME
                        Stop after the specified amount of time, e.g. (300s,
                        20m, 3h, 1h30m, etc.). Only used together with --no-
                        web
  -L LOGLEVEL, --loglevel=LOGLEVEL
                        Choose between DEBUG/INFO/WARNING/ERROR/CRITICAL.
                        Default is INFO.
  --logfile=LOGFILE     Path to log file. If not set, log will go to
                        stdout/stderr
  --print-stats         Print stats in the console
  --only-summary        Only print the summary stats
  --no-reset-stats      [DEPRECATED] Do not reset statistics once hatching has
                        been completed. This is now the default behavior. See
                        --reset-stats to disable
  --reset-stats         Reset statistics once hatching has been completed.
                        Should be set on both master and slaves when running
                        in distributed mode
  -l, --list            Show list of possible locust classes and exit
  --show-task-ratio     print table of the locust classes' task execution
                        ratio
  --show-task-ratio-json
                        print json data of the locust classes' task execution
                        ratio
  -V, --version         show program's version number and exit
  --exit-code-on-error=EXIT_CODE_ON_ERROR
                        sets the exit code to post on error
```
