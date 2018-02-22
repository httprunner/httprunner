## Development

To develop or debug `HttpRunner`, you shall clone source code first.

```bash
$ git clone https://github.com/HttpRunner/HttpRunner.git
```

Then install all dependencies:

```bash
$ pip install -r requirements-dev.txt
```

Now you can use `httprunner/cli.py` as debugging entrances.

```bash
# debug hrun
$ python httprunner/cli.py hrun -h

# debug locusts
$ python httprunner/cli.py locusts -h
```
