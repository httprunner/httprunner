#encoding: utf-8
import io
import os

from setuptools import find_packages, setup

about = {}
here = os.path.abspath(os.path.dirname(__file__))
with io.open(os.path.join(here, 'httprunner', '__about__.py'), encoding='utf-8') as f:
    exec(f.read(), about)

with io.open("README.rst", encoding='utf-8') as f:
    long_description = f.read()

install_requires = [
    "requests",
    "PyYAML",
    "Jinja2",
    "har2case",
    "colorama",
    "colorlog"
]

setup(
    name=about['__title__'],
    version=about['__version__'],
    description=about['__description__'],
    long_description=long_description,
    author=about['__author__'],
    author_email=about['__author_email__'],
    url=about['__url__'],
    license=about['__license__'],
    packages=find_packages(exclude=["examples", "tests", "tests.*"]),
    package_data={
        'httprunner': ["templates/*"],
    },
    keywords='api test',
    install_requires=install_requires,
    extras_require={
        'dev': ['flask']
    },
    classifiers=[
        "Development Status :: 3 - Alpha",
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3.4',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6'
    ],
    entry_points={
        'console_scripts': [
            'ate=httprunner.cli:main_hrun',
            'httprunner=httprunner.cli:main_hrun',
            'hrun=httprunner.cli:main_hrun',
            'locusts=httprunner.cli:main_locust'
        ]
    }
)
