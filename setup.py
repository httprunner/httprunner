#encoding: utf-8
import os
import re
from setuptools import setup, find_packages

# parse version from ate/__init__.py
with open(os.path.join(os.path.dirname(__file__), 'ate', '__init__.py')) as f:
    version = re.compile(r"__version__\s+=\s+'(.*)'", re.I).match(f.read()).group(1)

with open('README.md') as f:
    long_description = f.read()

setup(
    name='ApiTestEngine',
    version=version,
    description='An API test engine.',
    long_description=long_description,
    author='Leo Lee',
    author_email='mail@debugtalk.com',
    url='https://github.com/debugtalk/ApiTestEngine',
    license='MIT',
    packages=find_packages(exclude=['test.*', 'test']),
    keywords='api test',
    install_requires=[
        "requests",
        "termcolor",
        "flask",
        "PyYAML",
        "coveralls",
        "coverage",
        "PyUnitReport"
    ],
    dependency_links=[
        "git+https://github.com/debugtalk/PyUnitReport.git#egg=PyUnitReport"
    ],
    classifiers=[
        "Development Status :: 3 - Alpha",
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3.3',
        'Programming Language :: Python :: 3.4',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6'
    ],
    entry_points={
        'console_scripts': [
            'ate=ate.cli:main'
        ]
    }
)
