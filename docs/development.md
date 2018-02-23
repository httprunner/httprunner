## Development

To develop or debug `HttpRunner`, you shall clone source code first.

```bash
$ git clone https://github.com/HttpRunner/HttpRunner.git
```

Then install all dependencies:

```bash
$ pip install -r requirements-dev.txt
```

Now you can use `main-debug.py` as debugging entrances.

```bash
# debug hrun
$ python main-debug.py hrun -h

# debug locusts
$ python main-debug.py locusts -h
```
