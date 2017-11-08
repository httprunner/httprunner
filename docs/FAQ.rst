FAQ
===

Unable to install PyUnitReport dependency library automatically
---------------------------------------------------------------

If there is something goes wrong in installation like below. ::

    Downloading/unpacking PyUnitReport (from HttpRunner)
      Could not find any downloads that satisfy the requirement PyUnitReport (from HttpRunner)

You could install ``PyUnitReport`` manully at first. ::

    pip install PyUnitReport


And then everything will be OK when you reinstall ``HttpRunner``. ::

    pip install HttpRunner
