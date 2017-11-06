#encoding: utf-8
import io
import os
import re

from setuptools import find_packages, setup

# parse version from ate/__init__.py
with open(os.path.join(os.path.dirname(__file__), 'ate', '__init__.py')) as f:
    version = re.compile(r"__version__\s+=\s+'(.*)'", re.I).match(f.read()).group(1)

with io.open("README.md", encoding='utf-8') as f:
    long_description = f.read()

setup(
    name='HttpRunner',
    version=version,
    description='HTTP test runner, not just about api test and load test.',
    long_description=long_description,
    author='Leo Lee',
    author_email='mail@debugtalk.com',
    url='https://github.com/debugtalk/HttpRunner',
    license='MIT',
    packages=find_packages(exclude=['test.*', 'test']),
    package_data={
        'ate': ['locustfile_template'],
    },
    keywords='api test',
    install_requires=[
        "requests[security]",
        "flask",
        "PyYAML",
        "coveralls",
        "coverage",
        "PyUnitReport"
    ],
    extras_require={
        'locustio': [
            "locustio"
        ]
    },
    dependency_links=[
        "git+https://github.com/debugtalk/PyUnitReport.git#egg=PyUnitReport-0",
        "git+https://github.com/locustio/locust.git#egg=locust-0"
    ],
    classifiers=[
        "Development Status :: 3 - Alpha",
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3.4',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6'
    ],
    entry_points={
        'console_scripts': [
            'ate=ate.cli:main_ate',
            'locusts=ate.cli:main_locust'
        ]
    }
)
