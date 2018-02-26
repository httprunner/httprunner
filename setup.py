#encoding: utf-8
import io

from httprunner import __version__
from setuptools import find_packages, setup

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
    name='HttpRunner',
    version=__version__,
    description='HTTP test runner, not just about api test and load test.',
    long_description=long_description,
    author='Leo Lee',
    author_email='mail@debugtalk.com',
    url='https://github.com/HttpRunner/HttpRunner',
    license='MIT',
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
